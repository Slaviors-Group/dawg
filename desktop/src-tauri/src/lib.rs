use serde::{Deserialize, Serialize};
#[cfg(windows)]
use std::os::windows::process::CommandExt;
use std::path::{Path, PathBuf};
use std::process::Command;
use tauri::{AppHandle, Manager};

/// Windows CREATE_NO_WINDOW process creation flag. Without this, every engine
/// subprocess invocation pops up its own visible console host window (e.g.
/// Windows Terminal) since the Tauri desktop shell has no console of its own
/// to inherit.
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
        // If engine is at .../resources/binaries/dawg.exe -> resource dir is .../resources
        if parent.ends_with("binaries") {
            if let Some(grandparent) = parent.parent() {
                return Some(grandparent.to_path_buf());
            }
        }
        // If engine is at .../bin/dawg.exe in repo -> resource dir is ...
        if parent.ends_with("bin") {
            if let Some(grandparent) = parent.parent() {
                return Some(grandparent.to_path_buf());
            }
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
    if let Ok(out) = output {
        if out.status.success() {
            let text = String::from_utf8_lossy(&out.stdout);
            for line in text.lines() {
                if line.contains("REG_SZ") || line.contains("REG_EXPAND_SZ") {
                    let parts: Vec<&str> = line.split_whitespace().collect();
                    if parts.len() >= 3 {
                        // The PATH value may contain spaces and semicolons
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
    if let Ok(current_exe) = std::env::current_exe() {
        if let Some(parent) = current_exe.parent() {
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
                    let res = infer_resource_dir(candidate).unwrap_or_else(|| parent.to_path_buf());
                    return (candidate.clone(), Some(res), true);
                }
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
        if dev_path.exists() {
            if let Ok(abs) = dev_path.canonicalize() {
                let res_dir = infer_resource_dir(&abs);
                return (abs, res_dir, false);
            }
        }
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

fn build_engine_command(app: &AppHandle, subcommand: &str, args: &[String]) -> Command {
    let (engine_path, resource_dir, _) = resolve_engine_binary(app);
    let mut cmd = Command::new(&engine_path);
    suppress_console_window(&mut cmd);
    cmd.arg(subcommand);
    for arg in args {
        cmd.arg(arg);
    }
    cmd.arg("--output").arg("json");

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
async fn start_capture(app: AppHandle, url: String) -> Result<CommandOutput, String> {
    execute_engine_cmd(app, "capture".to_string(), vec!["--url".to_string(), url]).await
}

#[tauri::command]
async fn stop_capture(
    app: AppHandle,
    control_file: Option<String>,
) -> Result<CommandOutput, String> {
    let mut args = vec!["stop".to_string()];
    if let Some(cf) = control_file {
        args.push("--control-file".to_string());
        args.push(cf);
    }
    execute_engine_cmd(app, "capture".to_string(), args).await
}

#[tauri::command]
async fn inspect_artifact(app: AppHandle, path: String) -> Result<CommandOutput, String> {
    execute_engine_cmd(app, "inspect".to_string(), vec![path]).await
}

#[tauri::command]
async fn run_replay(app: AppHandle, artifact: String) -> Result<CommandOutput, String> {
    execute_engine_cmd(app, "run".to_string(), vec![artifact]).await
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
        .invoke_handler(tauri::generate_handler![
            check_engine_installed,
            get_doctor_report,
            start_capture,
            stop_capture,
            inspect_artifact,
            run_replay,
            verify_result
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
