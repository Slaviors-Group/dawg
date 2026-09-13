use serde::{Deserialize, Serialize};
#[cfg(windows)]
use std::os::windows::process::CommandExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::sync::Mutex;
use tauri::{AppHandle, Manager};

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

/// Tracks replay and capture processes for cancellation and exit cleanup.
#[derive(Default)]
struct ProcessRegistry {
    replay_pid: Mutex<Option<u32>>,
    replay_cancelled: Mutex<bool>,
    capture_daemon_pid: Mutex<Option<u32>>,
}

impl ProcessRegistry {
    fn start_replay(&self, pid: u32) {
        *self.replay_pid.lock().unwrap() = Some(pid);
        *self.replay_cancelled.lock().unwrap() = false;
    }

    /// Clears the tracked replay PID and returns whether it had been
    /// cancelled by the user before this call.
    fn finish_replay(&self) -> bool {
        *self.replay_pid.lock().unwrap() = None;
        let mut cancelled = self.replay_cancelled.lock().unwrap();
        let was_cancelled = *cancelled;
        *cancelled = false;
        was_cancelled
    }

    /// Force-kills the currently tracked replay process tree, if any.
    /// Returns true if a replay was actually running and got cancelled.
    fn cancel_replay(&self) -> bool {
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

    fn set_capture_daemon(&self, pid: u32) {
        *self.capture_daemon_pid.lock().unwrap() = Some(pid);
    }

    fn clear_capture_daemon(&self) {
        *self.capture_daemon_pid.lock().unwrap() = None;
    }

    /// Terminates all replay and capture processes tracked by this session.
    fn kill_all(&self) {
        if let Some(pid) = self.replay_pid.lock().unwrap().take() {
            kill_process_tree(pid);
        }
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
        let report: DoctorReport = serde_json::from_str(&stdout_str).map_err(|e| {
            format!(
                "Failed to decode doctor report: {}. Output: {}",
                e, stdout_str
            )
        })?;
        Ok(report)
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
) -> Result<CommandOutput, String> {
    let mut args = vec!["--url".to_string(), url];
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
async fn run_replay(
    app: AppHandle,
    registry: tauri::State<'_, ProcessRegistry>,
    artifact: String,
) -> Result<CommandOutput, String> {
    let mut cmd = build_engine_command(&app, "run", &[artifact]);
    cmd.stdin(Stdio::null());
    cmd.stdout(Stdio::piped());
    cmd.stderr(Stdio::piped());

    let child = cmd
        .spawn()
        .map_err(|e| format!("Failed to start replay: {}", e))?;
    let pid = child.id();
    registry.start_replay(pid);

    // Wait off the async executor thread: Child::wait_with_output blocks the
    // calling thread until the process exits (or is killed by cancel_replay),
    // which can take up to the engine's internal replay timeout.
    let wait_result = tauri::async_runtime::spawn_blocking(move || child.wait_with_output())
        .await
        .map_err(|e| format!("Replay wait task panicked: {}", e))?;

    let was_cancelled = registry.finish_replay();
    if was_cancelled {
        return Err("Replay cancelled by user.".to_string());
    }

    match wait_result {
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
async fn cancel_replay(registry: tauri::State<'_, ProcessRegistry>) -> Result<bool, String> {
    Ok(registry.cancel_replay())
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
            list_artifacts,
            import_artifact,
            export_artifact,
            run_replay,
            cancel_replay,
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
