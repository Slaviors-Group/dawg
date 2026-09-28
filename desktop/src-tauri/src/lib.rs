use serde::{Deserialize, Serialize};
use std::io::{BufRead, BufReader, Read, Write};
#[cfg(windows)]
use std::os::windows::process::CommandExt;
use std::path::{Path, PathBuf};
use std::process::{ChildStdin, Command, Stdio};
use std::sync::Mutex;
use std::thread;
use tauri::{AppHandle, Emitter, Manager};

/// Windows flag for starting engine processes without a console window.
#[cfg(windows)]
const CREATE_NO_WINDOW: u32 = 0x0800_0000;

/// Suppresses console window allocation for `cmd` on Windows. No-op elsewhere.
fn suppress_console_window(_cmd: &mut Command) {
    #[cfg(windows)]
    {
        _cmd.creation_flags(CREATE_NO_WINDOW);
    }
}

#[derive(Debug, Serialize, Deserialize)]
pub struct CommandOutput {
    pub status: String,
    pub payload: serde_json::Value,
}

#[derive(Clone, Copy, PartialEq, Eq)]
enum InteractiveProcess {
    Replay,
    Editor,
}

/// Tracks interactive browser and capture processes for cancellation and exit cleanup.
#[derive(Default)]
struct ProcessRegistry {
    active_interactive: Mutex<Option<InteractiveProcess>>,
    active_artifact: Mutex<Option<String>>,
    replay_pid: Mutex<Option<u32>>,
    replay_stdin: Mutex<Option<ChildStdin>>,
    replay_cancelled: Mutex<bool>,
    editor_pid: Mutex<Option<u32>>,
    editor_stdin: Mutex<Option<ChildStdin>>,
    editor_cancelled: Mutex<bool>,
    capture_daemon_pid: Mutex<Option<u32>>,
}

impl ProcessRegistry {
    fn reserve_interactive(&self, process: InteractiveProcess) -> Result<(), String> {
        let mut active = self.active_interactive.lock().unwrap();
        if active.is_some() {
            return Err("An interactive Replay or Editor is already running.".to_string());
        }
        *active = Some(process);
        Ok(())
    }

    fn release_interactive(&self, process: InteractiveProcess) {
        let mut active = self.active_interactive.lock().unwrap();
        if *active == Some(process) {
            *active = None;
        }
    }

    fn start_replay(&self, pid: u32, stdin: ChildStdin, artifact: String) {
        *self.active_artifact.lock().unwrap() = Some(artifact);
        *self.replay_pid.lock().unwrap() = Some(pid);
        *self.replay_stdin.lock().unwrap() = Some(stdin);
        *self.replay_cancelled.lock().unwrap() = false;
    }

    /// Clears the tracked replay PID and returns whether it had been
    /// cancelled by the user before this call.
    fn finish_replay(&self) -> bool {
        *self.active_artifact.lock().unwrap() = None;
        *self.replay_pid.lock().unwrap() = None;
        *self.replay_stdin.lock().unwrap() = None;
        let mut cancelled = self.replay_cancelled.lock().unwrap();
        let was_cancelled = *cancelled;
        *cancelled = false;
        self.release_interactive(InteractiveProcess::Replay);
        was_cancelled
    }

    fn start_editor(&self, pid: u32, stdin: ChildStdin, artifact: String) {
        *self.active_artifact.lock().unwrap() = Some(artifact);
        *self.editor_pid.lock().unwrap() = Some(pid);
        *self.editor_stdin.lock().unwrap() = Some(stdin);
        *self.editor_cancelled.lock().unwrap() = false;
    }

    fn finish_editor(&self) -> bool {
        *self.active_artifact.lock().unwrap() = None;
        *self.editor_pid.lock().unwrap() = None;
        *self.editor_stdin.lock().unwrap() = None;
        let mut cancelled = self.editor_cancelled.lock().unwrap();
        let was_cancelled = *cancelled;
        *cancelled = false;
        self.release_interactive(InteractiveProcess::Editor);
        was_cancelled
    }

    /// Force-kills the currently tracked replay process tree, if any.
    /// Returns true if a replay was actually running and got cancelled.
    fn cancel_replay(&self) -> bool {
        *self.replay_stdin.lock().unwrap() = None;
        let pid = *self.replay_pid.lock().unwrap();
        match pid {
            Some(pid) => {
                *self.replay_cancelled.lock().unwrap() = true;
                kill_process_tree(pid);
                true
            }
            None => false,
        }
    }

    fn control_replay(&self, command: serde_json::Value) -> Result<(), String> {
        self.write_interactive_command(&self.replay_stdin, command, "replay")
    }

    fn cancel_editor(&self) -> bool {
        *self.editor_stdin.lock().unwrap() = None;
        let pid = *self.editor_pid.lock().unwrap();
        match pid {
            Some(pid) => {
                *self.editor_cancelled.lock().unwrap() = true;
                kill_process_tree(pid);
                true
            }
            None => false,
        }
    }

    fn control_editor(&self, command: serde_json::Value) -> Result<(), String> {
        self.write_interactive_command(&self.editor_stdin, command, "editor")
    }

    fn write_interactive_command(
        &self,
        stdin_lock: &Mutex<Option<ChildStdin>>,
        command: serde_json::Value,
        process_name: &str,
    ) -> Result<(), String> {
        let mut stdin = stdin_lock.lock().unwrap();
        let Some(stdin) = stdin.as_mut() else {
            return Err(format!("No interactive {} is running.", process_name));
        };
        writeln!(stdin, "{}", command)
            .map_err(|error| format!("Failed to send {} control: {}", process_name, error))?;
        stdin
            .flush()
            .map_err(|error| format!("Failed to flush {} control: {}", process_name, error))
    }

    fn artifact_is_active(&self, artifact: &str) -> bool {
        self.active_artifact
            .lock()
            .unwrap()
            .as_deref()
            .is_some_and(|active| active == artifact)
    }

    fn set_capture_daemon(&self, pid: u32) {
        *self.capture_daemon_pid.lock().unwrap() = Some(pid);
    }

    fn clear_capture_daemon(&self) {
        *self.capture_daemon_pid.lock().unwrap() = None;
    }

    /// Terminates all replay and capture processes tracked by this session.
    fn kill_all(&self) {
        *self.replay_stdin.lock().unwrap() = None;
        *self.editor_stdin.lock().unwrap() = None;
        if let Some(pid) = self.replay_pid.lock().unwrap().take() {
            kill_process_tree(pid);
        }
        if let Some(pid) = self.editor_pid.lock().unwrap().take() {
            kill_process_tree(pid);
        }
        *self.active_interactive.lock().unwrap() = None;
        *self.active_artifact.lock().unwrap() = None;
        if let Some(pid) = self.capture_daemon_pid.lock().unwrap().take() {
            kill_process_tree(pid);
        }
    }
}

/// Terminates `pid` and its descendant processes.
fn kill_process_tree(pid: u32) {
    #[cfg(windows)]
    {
        let mut cmd = Command::new("taskkill");
        cmd.args(["/PID", &pid.to_string(), "/T", "/F"]);
        suppress_console_window(&mut cmd);
        let _ = cmd.output();
    }
    #[cfg(not(windows))]
    {
        // Best-effort: SIGKILL the process group first (covers detached
        // daemons/sandboxes that set up their own group), then the PID itself.
        let _ = Command::new("kill")
            .args(["-9", &format!("-{}", pid)])
            .output();
        let _ = Command::new("kill").args(["-9", &pid.to_string()]).output();
    }
}

#[derive(Debug, Serialize, Deserialize)]
pub struct EngineStatusInfo {
    pub installed: bool,
    pub version: Option<String>,
    pub path: Option<String>,
    pub bundled: bool,
    pub error: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ComponentStatus {
    pub name: String,
    pub installed: bool,
    pub path: Option<String>,
    pub version: Option<String>,
    pub bundled: bool,
    pub error: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DoctorReport {
    pub status: String,
    #[serde(rename = "enginePath")]
    pub engine_path: String,
    #[serde(rename = "resourceDir")]
    pub resource_dir: String,
    #[serde(rename = "isBundled")]
    pub is_bundled: bool,
    pub components: Vec<ComponentStatus>,
    #[serde(rename = "generatedAt")]
    pub generated_at: String,
    #[serde(default = "unknown_compatibility")]
    pub compatibility: CompatibilityReport,
}

fn unknown_compatibility() -> CompatibilityReport {
    CompatibilityReport {
        status: "unknown".to_string(),
        expected_application_version: String::new(),
        expected_desktop_version: String::new(),
        expected_extension_version: String::new(),
        detected_engine_version: None,
        current_schema_version: None,
        detected_extension_version: None,
        engine: "unknown".to_string(),
        schema: "unknown".to_string(),
        extension: "unknown".to_string(),
    }
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct CompatibilityReport {
    status: String,
    expected_application_version: String,
    expected_desktop_version: String,
    expected_extension_version: String,
    detected_engine_version: Option<String>,
    current_schema_version: Option<String>,
    detected_extension_version: Option<String>,
    engine: String,
    schema: String,
    extension: String,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
struct VersionContract {
    app_version: String,
    desktop_version: String,
    extension_version: String,
}

fn component_version(components: &[ComponentStatus], name: &str) -> Option<String> {
    components
        .iter()
        .find(|component| component.name == name)
        .and_then(|component| component.version.clone())
}

fn extension_display_version(components: &[ComponentStatus]) -> Option<String> {
    let extension = components
        .iter()
        .find(|component| component.name == "browser-extension")?;
    let path = extension.path.as_ref()?;
    let contents = std::fs::read_to_string(path).ok()?;
    serde_json::from_str::<serde_json::Value>(&contents)
        .ok()?
        .get("version_name")
        .and_then(|value| value.as_str())
        .map(str::to_owned)
        .or_else(|| extension.version.clone())
}

fn compatibility_status(expected: &str, detected: Option<&str>) -> String {
    match detected {
        Some(value) if value == expected => "compatible".to_string(),
        Some(_) => "mismatch".to_string(),
        None => "unknown".to_string(),
    }
}

fn apply_compatibility(mut report: DoctorReport) -> Result<DoctorReport, String> {
    let contract: VersionContract = serde_json::from_str(include_str!("../../../version.json"))
        .map_err(|error| format!("Invalid embedded version contract: {}", error))?;
    let engine = component_version(&report.components, "dawg-engine");
    let schema = component_version(&report.components, "schema:manifest");
    let extension = extension_display_version(&report.components);
    let engine_status = compatibility_status(&contract.app_version, engine.as_deref());
    let schema_status = compatibility_status(&contract.app_version, schema.as_deref());
    let extension_status = compatibility_status(&contract.extension_version, extension.as_deref());
    let status = if engine_status == "mismatch"
        || schema_status == "mismatch"
        || extension_status == "mismatch"
    {
        "mismatch"
    } else if engine_status == "compatible"
        && schema_status == "compatible"
        && extension_status == "compatible"
    {
        "compatible"
    } else {
        "unknown"
    };
    report.compatibility = CompatibilityReport {
        status: status.to_string(),
        expected_application_version: contract.app_version,
        expected_desktop_version: contract.desktop_version,
        expected_extension_version: contract.extension_version,
        detected_engine_version: engine,
        current_schema_version: schema,
        detected_extension_version: extension,
        engine: engine_status,
        schema: schema_status,
        extension: extension_status,
    };
    Ok(report)
}

/// Helper to determine the resource root given a resolved binary path.
fn infer_resource_dir(engine_path: &Path) -> Option<PathBuf> {
    if let Some(parent) = engine_path.parent() {
        if parent.ends_with("binaries") || parent.ends_with("bin") {
            return parent.parent().map(Path::to_path_buf);
        }
        // Check if there is a sibling "resources", "scripts", or "schema" folder
        if parent.join("scripts").exists() || parent.join("schema").exists() {
            return Some(parent.to_path_buf());
        }
        if parent.join("resources").exists() {
            return Some(parent.join("resources"));
        }
    }
    None
}

#[cfg(windows)]
fn query_registry_path(hive: &str, subkey: &str) -> Vec<PathBuf> {
    let mut reg_cmd = Command::new("reg");
    reg_cmd.args(["query", &format!("{}\\{}", hive, subkey), "/v", "Path"]);
    suppress_console_window(&mut reg_cmd);
    let output = reg_cmd.output();

    let mut paths = Vec::new();
    let Ok(out) = output else {
        return paths;
    };
    if out.status.success() {
        let text = String::from_utf8_lossy(&out.stdout);
        for line in text.lines() {
            if line.contains("REG_SZ") || line.contains("REG_EXPAND_SZ") {
                let parts: Vec<&str> = line.split_whitespace().collect();
                if parts.len() >= 3 {
                    let path_val = parts[2..].join(" ");
                    for p in path_val.split(';') {
                        let trimmed = p.trim();
                        if !trimmed.is_empty() {
                            paths.push(PathBuf::from(trimmed));
                        }
                    }
                }
            }
        }
    }
    paths
}

/// Resolves the engine executable path and resource directory:
/// 1. Bundled Tauri resource directory (all nested/flat permutations)
/// 2. Executable parent directory (installed app root)
/// 3. Common user install locations (%LOCALAPPDATA%, %PROGRAMFILES%)
/// 4. Monorepo development paths (`../bin/dawg.exe`)
/// 5. Realtime Windows Registry User & System PATH
/// 6. In-memory process PATH fallback
fn resolve_engine_binary(app: &AppHandle) -> (PathBuf, Option<PathBuf>, bool) {
    let exe_name = if cfg!(windows) { "dawg.exe" } else { "dawg" };

    // 1. Check Tauri Resource directory (covers both flat and nested packaging)
    if let Ok(resource_dir) = app.path().resource_dir() {
        let candidates = [
            resource_dir
                .join("resources")
                .join("binaries")
                .join(exe_name),
            resource_dir.join("binaries").join(exe_name),
            resource_dir.join("resources").join(exe_name),
            resource_dir.join(exe_name),
        ];
        for candidate in &candidates {
            if candidate.exists() {
                let res = infer_resource_dir(candidate).unwrap_or(resource_dir);
                return (candidate.clone(), Some(res), true);
            }
        }
    }

    // 2. Check sibling of current desktop executable
    if let Some(parent) = std::env::current_exe()
        .ok()
        .and_then(|path| path.parent().map(Path::to_path_buf))
    {
        let candidates = [
            parent
                .join("resources")
                .join("resources")
                .join("binaries")
                .join(exe_name),
            parent.join("resources").join("binaries").join(exe_name),
            parent.join("binaries").join(exe_name),
            parent.join("resources").join(exe_name),
            parent.join(exe_name),
        ];
        for candidate in &candidates {
            if candidate.exists() {
                let res = infer_resource_dir(candidate).unwrap_or_else(|| parent.clone());
                return (candidate.clone(), Some(res), true);
            }
        }
    }

    // 3. Check common Windows install directories
    #[cfg(windows)]
    {
        let mut install_candidates = Vec::new();
        if let Ok(local_app_data) = std::env::var("LOCALAPPDATA") {
            let base = PathBuf::from(local_app_data).join("Programs").join("DAWG");
            install_candidates.push(base.join("resources").join("binaries").join(exe_name));
            install_candidates.push(base.join("binaries").join(exe_name));
            install_candidates.push(base.join(exe_name));
        }
        if let Ok(program_files) = std::env::var("ProgramFiles") {
            let base = PathBuf::from(program_files).join("DAWG");
            install_candidates.push(base.join("resources").join("binaries").join(exe_name));
            install_candidates.push(base.join("binaries").join(exe_name));
            install_candidates.push(base.join(exe_name));
        }
        if let Ok(user_profile) = std::env::var("USERPROFILE") {
            let base = PathBuf::from(user_profile).join(".dawg").join("bin");
            install_candidates.push(base.join(exe_name));
        }
        for candidate in install_candidates {
            if candidate.exists() {
                let res = infer_resource_dir(&candidate);
                return (candidate, res, true);
            }
        }
    }

    // 4. Check monorepo development paths
    let dev_candidates = [
        PathBuf::from("../bin").join(exe_name),
        PathBuf::from("../../bin").join(exe_name),
        PathBuf::from("bin").join(exe_name),
        PathBuf::from("../engine/bin").join(exe_name),
    ];
    for dev_path in &dev_candidates {
        if !dev_path.exists() {
            continue;
        }
        let Ok(abs) = dev_path.canonicalize() else {
            continue;
        };
        let res_dir = infer_resource_dir(&abs);
        return (abs, res_dir, false);
    }

    // 5. Query Windows Registry for real-time updated User/System PATH
    #[cfg(windows)]
    {
        let mut reg_dirs = query_registry_path("HKCU", "Environment");
        reg_dirs.extend(query_registry_path(
            "HKLM",
            "SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Environment",
        ));
        for dir in reg_dirs {
            let target = dir.join(exe_name);
            if target.exists() {
                let res = infer_resource_dir(&target);
                return (target, res, false);
            }
        }
    }

    // 6. Fallback to process in-memory PATH
    if let Ok(path_var) = std::env::var("PATH") {
        for dir in std::env::split_paths(&path_var) {
            let target = dir.join(exe_name);
            if target.exists() {
                let res = infer_resource_dir(&target);
                return (target, res, false);
            }
        }
    }

    (PathBuf::from(exe_name), None, false)
}

fn has_output_argument(args: &[String]) -> bool {
    args.iter()
        .any(|arg| arg == "--output" || arg.starts_with("--output="))
}

fn build_engine_command(app: &AppHandle, subcommand: &str, args: &[String]) -> Command {
    let (engine_path, resource_dir, _) = resolve_engine_binary(app);
    let mut cmd = Command::new(&engine_path);
    suppress_console_window(&mut cmd);
    cmd.arg(subcommand);
    for arg in args {
        cmd.arg(arg);
    }
    // Some engine subcommands use --output as a destination path rather than
    // the global response format. Never overwrite that path with "json".
    if !has_output_argument(args) {
        cmd.arg("--output").arg("json");
    }

    if let Some(res_dir) = resource_dir {
        cmd.env("DAWG_RESOURCES_DIR", res_dir);
    }

    cmd
}

#[tauri::command]
async fn check_engine_installed(app: AppHandle) -> Result<EngineStatusInfo, String> {
    let (engine_path, resource_dir, is_bundled) = resolve_engine_binary(&app);
    let mut cmd = Command::new(&engine_path);
    suppress_console_window(&mut cmd);
    cmd.arg("--version");
    if let Some(ref res_dir) = resource_dir {
        cmd.env("DAWG_RESOURCES_DIR", res_dir);
    }

    match cmd.output() {
        Ok(output) => {
            if output.status.success() {
                let ver_str = String::from_utf8_lossy(&output.stdout).trim().to_string();
                Ok(EngineStatusInfo {
                    installed: true,
                    version: Some(ver_str),
                    path: Some(engine_path.to_string_lossy().to_string()),
                    bundled: is_bundled,
                    error: None,
                })
            } else {
                let err_str = String::from_utf8_lossy(&output.stderr).trim().to_string();
                Ok(EngineStatusInfo {
                    installed: false,
                    version: None,
                    path: Some(engine_path.to_string_lossy().to_string()),
                    bundled: is_bundled,
                    error: Some(err_str),
                })
            }
        }
        Err(err) => Ok(EngineStatusInfo {
            installed: false,
            version: None,
            path: Some(engine_path.to_string_lossy().to_string()),
            bundled: is_bundled,
            error: Some(format!("Failed to execute 'dawg' engine: {}", err)),
        }),
    }
}

#[tauri::command]
async fn get_doctor_report(app: AppHandle) -> Result<DoctorReport, String> {
    let mut cmd = build_engine_command(&app, "doctor", &[]);
    let output = cmd
        .output()
        .map_err(|e| format!("Failed to run 'dawg doctor': {}", e))?;
    let stdout_str = String::from_utf8_lossy(&output.stdout);

    if output.status.success() {
        let mut report: DoctorReport = serde_json::from_str(&stdout_str).map_err(|e| {
            format!(
                "Failed to decode doctor report: {}. Output: {}",
                e, stdout_str
            )
        })?;
        report.compatibility = unknown_compatibility();
        apply_compatibility(report)
    } else {
        let stderr_str = String::from_utf8_lossy(&output.stderr);
        Err(format!("Doctor check failed: {}", stderr_str.trim()))
    }
}

#[tauri::command]
async fn execute_engine_cmd(
    app: AppHandle,
    subcommand: String,
    args: Vec<String>,
) -> Result<CommandOutput, String> {
    let mut cmd = build_engine_command(&app, &subcommand, &args);

    match cmd.output() {
        Ok(output) => {
            let stdout_str = String::from_utf8_lossy(&output.stdout);
            if output.status.success() {
                let parsed: serde_json::Value = serde_json::from_str(&stdout_str)
                    .unwrap_or_else(|_| serde_json::json!({ "raw": stdout_str.trim() }));
                Ok(CommandOutput {
                    status: "success".to_string(),
                    payload: parsed,
                })
            } else {
                let stderr_str = String::from_utf8_lossy(&output.stderr);
                Err(format!("Engine command failed: {}", stderr_str.trim()))
            }
        }
        Err(err) => Err(format!("Failed to execute engine: {}", err)),
    }
}

#[tauri::command]
async fn start_capture(
    app: AppHandle,
    registry: tauri::State<'_, ProcessRegistry>,
    url: String,
    title: Option<String>,
    diagnostics_profile: Option<String>,
) -> Result<CommandOutput, String> {
    let profile = diagnostics_profile.unwrap_or_else(|| "safe".to_string());
    if profile != "safe" && profile != "enhanced" {
        return Err("Invalid diagnostics profile; expected safe or enhanced".to_string());
    }
    let mut args = vec![
        "--url".to_string(),
        url,
        "--diagnostics-profile".to_string(),
        profile,
    ];
    if let Some(title) = title.filter(|value| !value.trim().is_empty()) {
        args.push("--title".to_string());
        args.push(title);
    }
    let result = execute_engine_cmd(app, "capture".to_string(), args).await?;
    // Track the detached daemon PID so it can be force-killed on app exit
    // even though the short-lived launcher process above has already exited.
    if let Some(daemon_pid) = result.payload.get("daemonPid").and_then(|v| v.as_u64()) {
        registry.set_capture_daemon(daemon_pid as u32);
    }
    Ok(result)
}

#[tauri::command]
async fn stop_capture(
    app: AppHandle,
    registry: tauri::State<'_, ProcessRegistry>,
    control_file: Option<String>,
) -> Result<CommandOutput, String> {
    let mut args = vec!["stop".to_string()];
    if let Some(cf) = control_file {
        args.push("--control-file".to_string());
        args.push(cf);
    }
    let result = execute_engine_cmd(app, "capture".to_string(), args).await;
    registry.clear_capture_daemon();
    result
}

#[tauri::command]
async fn inspect_artifact(app: AppHandle, path: String) -> Result<CommandOutput, String> {
    execute_engine_cmd(app, "inspect".to_string(), vec![path]).await
}

#[tauri::command]
async fn inspect_diagnostics(app: AppHandle, artifact: String) -> Result<CommandOutput, String> {
    execute_engine_cmd(
        app,
        "diagnostics".to_string(),
        vec!["inspect".to_string(), artifact],
    )
    .await
}

#[tauri::command]
async fn export_diagnostics_har(
    app: AppHandle,
    artifact: String,
    output: String,
) -> Result<CommandOutput, String> {
    execute_engine_cmd(
        app,
        "diagnostics".to_string(),
        vec![
            "export-har".to_string(),
            artifact,
            "--output".to_string(),
            output,
        ],
    )
    .await
}

#[tauri::command]
async fn copy_diagnostics_curl(
    app: AppHandle,
    artifact: String,
    request_id: String,
) -> Result<CommandOutput, String> {
    execute_engine_cmd(
        app,
        "diagnostics".to_string(),
        vec![
            "copy-curl".to_string(),
            artifact,
            "--request-id".to_string(),
            request_id,
        ],
    )
    .await
}

#[tauri::command]
async fn remove_diagnostics(
    app: AppHandle,
    artifact: String,
    output_dir: String,
    categories: Vec<String>,
    body_refs: Vec<String>,
) -> Result<CommandOutput, String> {
    let mut args = vec![
        "remove".to_string(),
        artifact,
        "--output-dir".to_string(),
        output_dir,
    ];
    for category in categories {
        args.push("--remove-category".to_string());
        args.push(category);
    }
    for body_ref in body_refs {
        args.push("--remove-body-ref".to_string());
        args.push(body_ref);
    }
    execute_engine_cmd(app, "diagnostics".to_string(), args).await
}

#[tauri::command]
async fn list_artifacts(app: AppHandle) -> Result<CommandOutput, String> {
    execute_engine_cmd(app, "artifacts".to_string(), vec!["list".to_string()]).await
}

#[tauri::command]
async fn import_artifact(app: AppHandle, archive: String) -> Result<CommandOutput, String> {
    execute_engine_cmd(
        app,
        "artifacts".to_string(),
        vec!["import".to_string(), archive],
    )
    .await
}

#[tauri::command]
async fn export_artifact(
    app: AppHandle,
    artifact: String,
    output: String,
) -> Result<CommandOutput, String> {
    execute_engine_cmd(
        app,
        "artifacts".to_string(),
        vec![
            "export".to_string(),
            artifact,
            "--output".to_string(),
            output,
            "--force".to_string(),
        ],
    )
    .await
}

#[tauri::command]
async fn delete_artifact(
    app: AppHandle,
    registry: tauri::State<'_, ProcessRegistry>,
    artifact: String,
) -> Result<CommandOutput, String> {
    if registry.artifact_is_active(&artifact) {
        return Err("Stop Replay or Editor before deleting its active artifact.".to_string());
    }
    execute_engine_cmd(
        app,
        "artifacts".to_string(),
        vec!["delete".to_string(), artifact],
    )
    .await
}

#[tauri::command]
async fn review_artifact(
    app: AppHandle,
    artifact: String,
    review_file: String,
) -> Result<CommandOutput, String> {
    execute_engine_cmd(
        app,
        "artifacts".to_string(),
        vec![
            "review".to_string(),
            artifact,
            "--review-file".to_string(),
            review_file,
        ],
    )
    .await
}

const REPLAY_EVENT_PREFIX: &str = "DAWG_REPLAY_EVENT\t";
const EDITOR_EVENT_PREFIX: &str = "DAWG_EDITOR_EVENT\t";

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct ReplayControl {
    id: String,
    #[serde(rename = "type")]
    kind: String,
    offset_ms: Option<u64>,
    speed: Option<f64>,
}

#[tauri::command]
async fn run_replay(
    app: AppHandle,
    registry: tauri::State<'_, ProcessRegistry>,
    artifact: String,
) -> Result<CommandOutput, String> {
    // Desktop replays are intentionally interactive; direct CLI `dawg run`
    // remains the finite screenshot-producing workflow.
    registry.reserve_interactive(InteractiveProcess::Replay)?;
    let mut cmd = build_engine_command(
        &app,
        "run",
        &["--interactive".to_string(), artifact.clone()],
    );
    cmd.stdin(Stdio::piped());
    cmd.stdout(Stdio::piped());
    cmd.stderr(Stdio::piped());

    let mut child = match cmd.spawn() {
        Ok(child) => child,
        Err(error) => {
            registry.release_interactive(InteractiveProcess::Replay);
            return Err(format!("Failed to start replay: {}", error));
        }
    };
    let pid = child.id();
    let stdin = match child.stdin.take() {
        Some(stdin) => stdin,
        None => {
            kill_process_tree(pid);
            registry.release_interactive(InteractiveProcess::Replay);
            return Err("Failed to open interactive replay control channel.".to_string());
        }
    };
    registry.start_replay(pid, stdin, artifact.clone());
    let stdout = match child.stdout.take() {
        Some(stdout) => stdout,
        None => {
            kill_process_tree(pid);
            registry.finish_replay();
            return Err("Failed to open replay result channel.".to_string());
        }
    };
    let stderr = match child.stderr.take() {
        Some(stderr) => stderr,
        None => {
            kill_process_tree(pid);
            registry.finish_replay();
            return Err("Failed to open replay event channel.".to_string());
        }
    };
    let event_app = app.clone();

    // Drain stdout and stderr concurrently. Replay state is streamed on stderr
    // while stdout remains the engine's final machine-readable result.
    let wait_result = tauri::async_runtime::spawn_blocking(move || {
        let stdout_reader = thread::spawn(move || {
            let mut contents = String::new();
            let result = BufReader::new(stdout).read_to_string(&mut contents);
            (contents, result)
        });
        let stderr_reader = thread::spawn(move || {
            let mut diagnostics = Vec::new();
            for line in BufReader::new(stderr).lines() {
                let line = line?;
                if let Some(payload) = line.strip_prefix(REPLAY_EVENT_PREFIX) {
                    match serde_json::from_str::<serde_json::Value>(payload) {
                        Ok(event) => {
                            let _ = event_app.emit("dawg://replay-event", event);
                        }
                        Err(error) => diagnostics.push(format!(
                            "Invalid replay event from engine: {} ({})",
                            payload, error
                        )),
                    }
                } else {
                    diagnostics.push(line);
                }
            }
            Ok::<String, std::io::Error>(diagnostics.join("\n"))
        });
        let status = child.wait();
        let (stdout, stdout_result) = stdout_reader
            .join()
            .map_err(|_| "Replay stdout reader panicked".to_string())?;
        stdout_result.map_err(|error| format!("Failed to read replay result: {}", error))?;
        let stderr = stderr_reader
            .join()
            .map_err(|_| "Replay stderr reader panicked".to_string())?
            .map_err(|error| format!("Failed to read replay events: {}", error))?;
        status
            .map(|status| (status, stdout, stderr))
            .map_err(|error| format!("Failed to wait for replay: {}", error))
    })
    .await
    .map_err(|e| format!("Replay wait task panicked: {}", e))?;

    let was_cancelled = registry.finish_replay();
    if was_cancelled {
        return Err("Replay cancelled by user.".to_string());
    }

    match wait_result {
        Ok((status, stdout, stderr)) => {
            if status.success() {
                let parsed: serde_json::Value = serde_json::from_str(&stdout)
                    .unwrap_or_else(|_| serde_json::json!({ "raw": stdout.trim() }));
                Ok(CommandOutput {
                    status: "success".to_string(),
                    payload: parsed,
                })
            } else {
                Err(format!("Engine command failed: {}", stderr.trim()))
            }
        }
        Err(err) => Err(format!("Failed to execute engine: {}", err)),
    }
}

#[tauri::command]
fn startup_artifact() -> Option<String> {
    std::env::args_os()
        .skip(1)
        .map(PathBuf::from)
        .find(|path| {
            path.is_file()
                && path
                    .extension()
                    .is_some_and(|extension| extension.eq_ignore_ascii_case("dawg"))
        })
        .map(|path| path.to_string_lossy().into_owned())
}

#[tauri::command]
async fn cancel_replay(registry: tauri::State<'_, ProcessRegistry>) -> Result<bool, String> {
    Ok(registry.cancel_replay())
}

#[tauri::command]
async fn control_replay(
    registry: tauri::State<'_, ProcessRegistry>,
    command: ReplayControl,
) -> Result<(), String> {
    if command.id.trim().is_empty() {
        return Err("Replay controls require a command ID.".to_string());
    }
    match command.kind.as_str() {
        "play" | "pause" | "getState" => {}
        "seek" if command.offset_ms.is_some() => {}
        "setSpeed"
            if command
                .speed
                .is_some_and(|speed| matches!(speed, 0.5 | 1.0 | 1.5 | 2.0 | 4.0)) => {}
        "seek" => return Err("Replay seek requires offsetMs.".to_string()),
        "setSpeed" => return Err("Unsupported replay speed.".to_string()),
        _ => return Err("Unsupported replay control command.".to_string()),
    }
    let mut payload = serde_json::json!({
        "protocol": "dawg.replay.v1",
        "id": command.id,
        "type": command.kind,
    });
    if let Some(offset_ms) = command.offset_ms {
        payload["offsetMs"] = serde_json::json!(offset_ms);
    }
    if let Some(speed) = command.speed {
        payload["speed"] = serde_json::json!(speed);
    }
    registry.control_replay(payload)
}

#[tauri::command]
async fn seek_replay(
    registry: tauri::State<'_, ProcessRegistry>,
    offset_ms: u64,
) -> Result<(), String> {
    registry.control_replay(serde_json::json!({
        "protocol": "dawg.replay.v1",
        "id": format!("diagnostic-seek-{}", offset_ms),
        "type": "seek",
        "offsetMs": offset_ms,
    }))
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct EditorControl {
    id: String,
    #[serde(rename = "type")]
    kind: String,
    offset_ms: Option<u64>,
    speed: Option<f64>,
    flag: Option<serde_json::Value>,
    flag_id: Option<String>,
    flags: Option<serde_json::Value>,
    artifact_title: Option<String>,
}

#[tauri::command]
async fn launch_editor(
    app: AppHandle,
    registry: tauri::State<'_, ProcessRegistry>,
    artifact: String,
) -> Result<CommandOutput, String> {
    registry.reserve_interactive(InteractiveProcess::Editor)?;
    let mut cmd = build_engine_command(
        &app,
        "editor",
        &[artifact.clone(), "--interactive".to_string()],
    );
    cmd.stdin(Stdio::piped());
    cmd.stdout(Stdio::piped());
    cmd.stderr(Stdio::piped());

    let mut child = match cmd.spawn() {
        Ok(child) => child,
        Err(error) => {
            registry.release_interactive(InteractiveProcess::Editor);
            return Err(format!("Failed to start editor: {}", error));
        }
    };
    let pid = child.id();
    let stdin = match child.stdin.take() {
        Some(stdin) => stdin,
        None => {
            kill_process_tree(pid);
            registry.release_interactive(InteractiveProcess::Editor);
            return Err("Failed to open interactive editor control channel.".to_string());
        }
    };
    registry.start_editor(pid, stdin, artifact.clone());
    let stdout = match child.stdout.take() {
        Some(stdout) => stdout,
        None => {
            kill_process_tree(pid);
            registry.finish_editor();
            return Err("Failed to open editor result channel.".to_string());
        }
    };
    let stderr = match child.stderr.take() {
        Some(stderr) => stderr,
        None => {
            kill_process_tree(pid);
            registry.finish_editor();
            return Err("Failed to open editor event channel.".to_string());
        }
    };
    let event_app = app.clone();
    let wait_result = tauri::async_runtime::spawn_blocking(move || {
        let stdout_reader = thread::spawn(move || {
            let mut contents = String::new();
            let result = BufReader::new(stdout).read_to_string(&mut contents);
            (contents, result)
        });
        let stderr_reader = thread::spawn(move || {
            let mut diagnostics = Vec::new();
            for line in BufReader::new(stderr).lines() {
                let line = line?;
                if let Some(payload) = line.strip_prefix(EDITOR_EVENT_PREFIX) {
                    match serde_json::from_str::<serde_json::Value>(payload) {
                        Ok(event) => {
                            let event_type = event.get("type").and_then(|value| value.as_str());
                            let command_id = event
                                .get("sequence")
                                .and_then(|value| value.as_u64())
                                .map(|sequence| format!("publication-{}", sequence))
                                .unwrap_or_else(|| "publication-result".to_string());
                            let browser_command = match event_type {
                                Some("artifactSaved") => Some(serde_json::json!({
                                    "protocol": "dawg.editor.v1",
                                    "id": command_id,
                                    "type": "publicationConfirmed",
                                    "artifactTitle": event.get("artifactTitle").and_then(|value| value.as_str()).unwrap_or(""),
                                })),
                                Some("validationError") if event.get("review").is_some() => {
                                    Some(serde_json::json!({
                                        "protocol": "dawg.editor.v1",
                                        "id": command_id,
                                        "type": "publicationFailed",
                                        "message": event.get("message").and_then(|value| value.as_str()).unwrap_or("Unknown publication error"),
                                    }))
                                }
                                _ => None,
                            };
                            let _ = event_app.emit("dawg://editor-event", event);
                            if let Some(command) = browser_command {
                                let _ = event_app.state::<ProcessRegistry>().control_editor(command);
                            }
                        }
                        Err(error) => diagnostics.push(format!(
                            "Invalid editor event from engine: {} ({})",
                            payload, error
                        )),
                    }
                } else {
                    diagnostics.push(line);
                }
            }
            Ok::<String, std::io::Error>(diagnostics.join("\n"))
        });
        let status = child.wait();
        let (stdout, stdout_result) = stdout_reader
            .join()
            .map_err(|_| "Editor stdout reader panicked".to_string())?;
        stdout_result.map_err(|error| format!("Failed to read editor result: {}", error))?;
        let stderr = stderr_reader
            .join()
            .map_err(|_| "Editor stderr reader panicked".to_string())?
            .map_err(|error| format!("Failed to read editor events: {}", error))?;
        status
            .map(|status| (status, stdout, stderr))
            .map_err(|error| format!("Failed to wait for editor: {}", error))
    })
    .await
    .map_err(|error| format!("Editor wait task panicked: {}", error))?;

    let was_cancelled = registry.finish_editor();
    if was_cancelled {
        return Err("Editor cancelled by user.".to_string());
    }
    match wait_result {
        Ok((status, stdout, _stderr)) if status.success() => {
            let parsed: serde_json::Value = serde_json::from_str(&stdout)
                .unwrap_or_else(|_| serde_json::json!({ "raw": stdout.trim() }));
            Ok(CommandOutput {
                status: "success".to_string(),
                payload: parsed,
            })
        }
        Ok((_, _, stderr)) => Err(format!("Engine editor command failed: {}", stderr.trim())),
        Err(error) => Err(format!("Failed to execute editor: {}", error)),
    }
}

#[tauri::command]
async fn cancel_editor(registry: tauri::State<'_, ProcessRegistry>) -> Result<bool, String> {
    Ok(registry.cancel_editor())
}

#[tauri::command]
async fn control_editor(
    registry: tauri::State<'_, ProcessRegistry>,
    command: EditorControl,
) -> Result<(), String> {
    if command.id.trim().is_empty() {
        return Err("Editor controls require a command ID.".to_string());
    }
    match command.kind.as_str() {
        "play" | "pause" | "getState" | "discardDraft" | "close" => {}
        "saveDraft" | "saveArtifact" if command.flags.is_some() => {}
        "seek" if command.offset_ms.is_some() => {}
        "setSpeed"
            if command
                .speed
                .is_some_and(|speed| matches!(speed, 0.5 | 1.0 | 1.5 | 2.0 | 4.0)) => {}
        "addFlag" | "updateFlag" if command.flag.is_some() => {}
        "deleteFlag"
            if command
                .flag_id
                .as_deref()
                .is_some_and(|id| !id.trim().is_empty()) => {}
        "seek" => return Err("Editor seek requires offsetMs.".to_string()),
        "setSpeed" => return Err("Unsupported editor speed.".to_string()),
        "addFlag" | "updateFlag" => {
            return Err("Editor flag update requires flag data.".to_string());
        }
        "saveDraft" | "saveArtifact" => {
            return Err("Editor save requires review flags.".to_string());
        }
        "deleteFlag" => return Err("Editor delete requires flagId.".to_string()),
        _ => return Err("Unsupported editor control command.".to_string()),
    }
    let mut payload = serde_json::json!({
        "protocol": "dawg.editor.v1",
        "id": command.id,
        "type": command.kind,
    });
    if let Some(offset_ms) = command.offset_ms {
        payload["offsetMs"] = serde_json::json!(offset_ms);
    }
    if let Some(speed) = command.speed {
        payload["speed"] = serde_json::json!(speed);
    }
    if let Some(flag) = command.flag {
        payload["flag"] = flag;
    }
    if let Some(flag_id) = command.flag_id {
        payload["flagId"] = serde_json::json!(flag_id);
    }
    if let Some(flags) = command.flags {
        payload["flags"] = flags;
    }
    if let Some(artifact_title) = command.artifact_title {
        payload["artifactTitle"] = serde_json::json!(artifact_title);
    }
    registry.control_editor(payload)
}

#[tauri::command]
async fn verify_result(
    app: AppHandle,
    artifact: String,
    against: String,
) -> Result<CommandOutput, String> {
    execute_engine_cmd(
        app,
        "verify".to_string(),
        vec![artifact, "--against".to_string(), against],
    )
    .await
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_dialog::init())
        .manage(ProcessRegistry::default())
        .invoke_handler(tauri::generate_handler![
            check_engine_installed,
            get_doctor_report,
            start_capture,
            stop_capture,
            inspect_artifact,
            inspect_diagnostics,
            export_diagnostics_har,
            copy_diagnostics_curl,
            remove_diagnostics,
            list_artifacts,
            import_artifact,
            export_artifact,
            delete_artifact,
            review_artifact,
            startup_artifact,
            run_replay,
            cancel_replay,
            control_replay,
            seek_replay,
            launch_editor,
            cancel_editor,
            control_editor,
            verify_result
        ])
        .build(tauri::generate_context!())
        .expect("error while building tauri application")
        .run(|app_handle, event| {
            // Terminate tracked engine work when the application exits.
            if !matches!(event, tauri::RunEvent::ExitRequested { .. }) {
                return;
            }
            if let Some(registry) = app_handle.try_state::<ProcessRegistry>() {
                registry.kill_all();
            }
        });
}

#[cfg(test)]
mod tests {
    use super::has_output_argument;

    #[test]
    fn detects_separate_output_argument() {
        let args = vec![
            "export".to_string(),
            "artifact".to_string(),
            "--output".to_string(),
            "D:\\exports\\artifact.dawg".to_string(),
        ];

        assert!(has_output_argument(&args));
    }

    #[test]
    fn detects_equals_output_argument() {
        let args = vec!["--output=D:\\exports\\artifact.dawg".to_string()];

        assert!(has_output_argument(&args));
    }

    #[test]
    fn allows_json_format_when_no_output_argument_exists() {
        let args = vec!["list".to_string()];

        assert!(!has_output_argument(&args));
    }
}
