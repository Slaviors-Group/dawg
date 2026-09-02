use std::process::Command;
use serde::{Deserialize, Serialize};

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
    pub error: Option<String>,
}

#[tauri::command]
async fn check_engine_installed() -> Result<EngineStatusInfo, String> {
    match Command::new("dawg").arg("--version").output() {
        Ok(output) => {
            if output.status.success() {
                let ver_str = String::from_utf8_lossy(&output.stdout).trim().to_string();
                Ok(EngineStatusInfo {
                    installed: true,
                    version: Some(ver_str),
                    path: Some("dawg".to_string()),
                    error: None,
                })
            } else {
                let err_str = String::from_utf8_lossy(&output.stderr).trim().to_string();
                Ok(EngineStatusInfo {
                    installed: false,
                    version: None,
                    path: None,
                    error: Some(err_str),
                })
            }
        }
        Err(err) => Ok(EngineStatusInfo {
            installed: false,
            version: None,
            path: None,
            error: Some(format!("'dawg' binary not found on PATH: {}", err)),
        }),
    }
}

#[tauri::command]
async fn execute_engine_cmd(subcommand: String, args: Vec<String>) -> Result<CommandOutput, String> {
    let mut cmd = Command::new("dawg");
    cmd.arg(&subcommand);
    for arg in args {
        cmd.arg(arg);
    }
    cmd.arg("--output").arg("json");

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
        Err(err) => Err(format!("Failed to execute 'dawg' binary: {}", err)),
    }
}

#[tauri::command]
async fn start_capture(url: String) -> Result<CommandOutput, String> {
    execute_engine_cmd("capture".to_string(), vec!["--url".to_string(), url]).await
}

#[tauri::command]
async fn stop_capture(control_file: Option<String>) -> Result<CommandOutput, String> {
    let mut args = vec!["stop".to_string()];
    if let Some(cf) = control_file {
        args.push("--control-file".to_string());
        args.push(cf);
    }
    execute_engine_cmd("capture".to_string(), args).await
}

#[tauri::command]
async fn inspect_artifact(path: String) -> Result<CommandOutput, String> {
    execute_engine_cmd("inspect".to_string(), vec![path]).await
}

#[tauri::command]
async fn run_replay(artifact: String) -> Result<CommandOutput, String> {
    execute_engine_cmd("run".to_string(), vec![artifact]).await
}

#[tauri::command]
async fn verify_result(artifact: String, against: String) -> Result<CommandOutput, String> {
    execute_engine_cmd("verify".to_string(), vec![artifact, "--against".to_string(), against]).await
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .invoke_handler(tauri::generate_handler![
            check_engine_installed,
            start_capture,
            stop_capture,
            inspect_artifact,
            run_replay,
            verify_result
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
