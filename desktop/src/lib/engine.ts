import { invoke } from "@tauri-apps/api/core";

/**
 * Typed IPC bridge for DAWG Engine CLI subprocess invocations.
 * Implements IPC contract table from architecture.md §Desktop ↔ Engine IPC Contract.
 */

export interface ComponentStatus {
  name: string;
  installed: boolean;
  path?: string;
  version?: string;
  bundled: boolean;
  error?: string;
}

export interface DoctorReport {
  status: "ready" | "degraded";
  enginePath: string;
  resourceDir: string;
  isBundled: boolean;
  components: ComponentStatus[];
  generatedAt: string;
}

export interface EngineStatusInfo {
  installed: boolean;
  version?: string;
  path?: string;
  bundled: boolean;
  error?: string;
}

export interface StartCaptureOptions {
  url: string;
  title?: string;
}

export interface StartCaptureResult {
  status: "capturing";
  sessionId: string;
  sessionPath: string;
  controlFile: string;
  resultFile: string;
  daemonPid: number;
}

export interface StopCaptureOptions {
  controlFile?: string;
}

export interface StopCaptureResult {
  status: "packaged";
  controlFile: string;
  sessionId?: string;
  artifactPath?: string;
}

export interface ArtifactItem {
  id: string;
  path: string;
  targetUrl: string;
  createdAt: string;
  title: string;
  schemaVersion: string;
  origin: "captured" | "imported" | "pulled" | "legacy";
  status: "valid";
  components: string[];
}

export interface InspectOptions {
  path: string;
}

export interface InspectResult {
  id?: string;
  version?: string;
  schemaVersion?: string;
  title?: string;
  source?: Record<string, unknown>;
  layers?: Array<Record<string, unknown>>;
  sanitize?: Record<string, unknown>;
  determinism?: Record<string, unknown>;
  expectedOutcome?: Record<string, unknown>;
  [key: string]: unknown;
}

export interface RunReplayOptions {
  artifact: string;
}

export interface RunReplayResult {
  // Field names below intentionally match the engine's actual camelCase
  // JSON output (see dawgtypes.ReplayOutput / ReplayOutcomes in the Go
  // engine), not snake_case - the engine never emits snake_case keys.
  artifactId: string;
  status: "started" | "completed" | "failed";
  replayedAt: string;
  sandbox?: {
    composeProject: string;
    containerId: string;
  };
  outcomes?: {
    exitCode: number;
    screenshots?: string[];
    httpResponses?: string;
    /** Combined stdout+stderr from replay-browser.cjs, always populated
     * (even on success) so a visually blank replay can be diagnosed from
     * the Execution Logs without digging through temp files. */
    appLogs?: string;
  };
  [key: string]: unknown;
}

export interface VerifyOptions {
  artifact: string;
  against: string;
}

export interface VerifyCheck {
  type: string;
  passed: boolean;
  expected: unknown;
  actual: unknown;
  diff_pixels?: number;
  threshold?: number;
}

export interface VerifyResult {
  artifact_id: string;
  verified_against: string;
  result: "pass" | "fail";
  summary: string;
  checks: VerifyCheck[];
  [key: string]: unknown;
}

export interface CommandOutput<T = unknown> {
  status: "success" | "error";
  payload: T;
}

export class EngineBridge {
  async checkInstalled(): Promise<EngineStatusInfo> {
    return invoke<EngineStatusInfo>("check_engine_installed");
  }

  async getDoctorReport(): Promise<DoctorReport> {
    return invoke<DoctorReport>("get_doctor_report");
  }

  async startCapture(options: StartCaptureOptions): Promise<StartCaptureResult> {
    try {
      const res = await invoke<CommandOutput<StartCaptureResult>>("start_capture", {
        url: options.url,
        title: options.title,
      });
      return res.payload;
    } catch (err) {
      console.warn("Tauri engine IPC fallback:", err);
      throw new Error(`Engine startCapture failed: ${String(err)}`);
    }
  }

  async stopCapture(options?: StopCaptureOptions): Promise<StopCaptureResult> {
    try {
      const res = await invoke<CommandOutput<StopCaptureResult>>("stop_capture", {
        controlFile: options?.controlFile,
      });
      return res.payload;
    } catch (err) {
      console.warn("Tauri engine IPC fallback:", err);
      throw new Error(`Engine stopCapture failed: ${String(err)}`);
    }
  }

  async listArtifacts(): Promise<ArtifactItem[]> {
    try {
      const res = await invoke<CommandOutput<ArtifactItem[]>>("list_artifacts");
      return res.payload;
    } catch (err) {
      console.warn("Tauri engine IPC fallback:", err);
      throw new Error(`Engine listArtifacts failed: ${String(err)}`);
    }
  }

  async importArtifact(archive: string): Promise<ArtifactItem> {
    try {
      const res = await invoke<CommandOutput<ArtifactItem>>("import_artifact", { archive });
      return res.payload;
    } catch (err) {
      console.warn("Tauri engine IPC fallback:", err);
      throw new Error(`Engine importArtifact failed: ${String(err)}`);
    }
  }

  async exportArtifact(artifact: string, output: string): Promise<{ status: "exported"; output: string }> {
    try {
      const res = await invoke<CommandOutput<{ status: "exported"; output: string }>>(
        "export_artifact",
        { artifact, output },
      );
      return res.payload;
    } catch (err) {
      console.warn("Tauri engine IPC fallback:", err);
      throw new Error(`Engine exportArtifact failed: ${String(err)}`);
    }
  }

  async inspectArtifact(options: InspectOptions): Promise<InspectResult> {
    try {
      const res = await invoke<CommandOutput<InspectResult>>("inspect_artifact", {
        path: options.path,
      });
      return res.payload;
    } catch (err) {
      console.warn("Tauri engine IPC fallback:", err);
      throw new Error(`Engine inspectArtifact failed: ${String(err)}`);
    }
  }

  async runReplay(options: RunReplayOptions): Promise<RunReplayResult> {
    try {
      const res = await invoke<CommandOutput<RunReplayResult>>("run_replay", {
        artifact: options.artifact,
      });
      return res.payload;
    } catch (err) {
      console.warn("Tauri engine IPC fallback:", err);
      throw new Error(`Engine runReplay failed: ${String(err)}`);
    }
  }

  /**
   * Force-kills an in-flight `run_replay` invocation (and its underlying
   * node/Chromium/mitmdump process tree) started by this session. Returns
   * true if a replay was actually running and got cancelled.
   */
  async cancelReplay(): Promise<boolean> {
    try {
      return await invoke<boolean>("cancel_replay");
    } catch (err) {
      console.warn("Tauri engine IPC fallback:", err);
      throw new Error(`Engine cancelReplay failed: ${String(err)}`);
    }
  }

  async verifyResult(options: VerifyOptions): Promise<VerifyResult> {
    try {
      const res = await invoke<CommandOutput<VerifyResult>>("verify_result", {
        artifact: options.artifact,
        against: options.against,
      });
      return res.payload;
    } catch (err) {
      console.warn("Tauri engine IPC fallback:", err);
      throw new Error(`Engine verifyResult failed: ${String(err)}`);
    }
  }
}

export const engine = new EngineBridge();
