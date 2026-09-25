import { invoke } from "@tauri-apps/api/core";

/** Typed IPC bridge for DAWG engine subprocess commands. */

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

export type CaptureDiagnosticsProfile = "safe" | "enhanced";

export interface StartCaptureOptions {
  url: string;
  title?: string;
  /** Selected by the desktop UI; engine support is intentionally forward-compatible. */
  diagnosticsProfile: CaptureDiagnosticsProfile;
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

export interface ReplayTimeline {
  firstTimestamp: number;
  lastTimestamp: number;
  durationMs: number;
}

export interface DiagnosticEvidence {
  summary?: Record<string, unknown>;
  timeline?: ReplayTimeline;
  timelineState?: "available" | "unavailable" | "limit-exceeded";
  console: Array<Record<string, unknown>>;
  network: Array<Record<string, unknown>>;
  errors: Array<Record<string, unknown>>;
  device?: Record<string, unknown>;
  bodies: Record<string, string>;
}

export type DiagnosticCategory = "console" | "network" | "errors" | "device" | "bodies";

export interface RemoveDiagnosticsOptions {
  artifact: string;
  outputDir: string;
  categories: DiagnosticCategory[];
  bodyRefs: string[];
}

export interface PackagedArtifact {
  directory: string;
  manifestPath: string;
  ociManifestDigest: string;
}

export interface RunReplayOptions {
  artifact: string;
}

export type ReplaySpeed = 0.5 | 1 | 1.5 | 2 | 4;
export type ReplayControlType = "play" | "pause" | "seek" | "setSpeed" | "getState";

export interface ReplayControl {
  id: string;
  type: ReplayControlType;
  offsetMs?: number;
  speed?: ReplaySpeed;
}

export interface ReplayEvent {
  protocol: "dawg.replay.v1";
  sequence: number;
  type: "ready" | "state" | "ack" | "error" | "finished" | "closed";
  reason?: string;
  commandId?: string | null;
  command?: ReplayControlType;
  currentTimeMs?: number;
  durationMs?: number;
  firstTimestamp?: number;
  lastTimestamp?: number;
  playing?: boolean;
  speed?: ReplaySpeed;
  message?: string;
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
  diffPixels?: number;
  threshold?: number;
}

export interface VerifyResult {
  artifactId: string;
  verifiedAgainst: string;
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
        diagnosticsProfile: options.diagnosticsProfile,
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

  async exportArtifact(
    artifact: string,
    output: string,
  ): Promise<{ status: "exported"; output: string }> {
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

  async inspectDiagnostics(artifact: string): Promise<DiagnosticEvidence> {
    const res = await invoke<CommandOutput<DiagnosticEvidence>>("inspect_diagnostics", {
      artifact,
    });
    return res.payload;
  }

  async exportDiagnosticsHAR(artifact: string, output: string): Promise<void> {
    await invoke<CommandOutput<unknown>>("export_diagnostics_har", { artifact, output });
  }

  async copyDiagnosticsCurl(artifact: string, requestId: string): Promise<string> {
    const res = await invoke<CommandOutput<{ raw?: string }>>("copy_diagnostics_curl", {
      artifact,
      requestId,
    });
    return res.payload.raw ?? String(res.payload);
  }

  async removeDiagnostics(options: RemoveDiagnosticsOptions): Promise<PackagedArtifact> {
    const res = await invoke<CommandOutput<PackagedArtifact>>("remove_diagnostics", {
      artifact: options.artifact,
      outputDir: options.outputDir,
      categories: options.categories,
      bodyRefs: options.bodyRefs,
    });
    return res.payload;
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

  async controlReplay(command: ReplayControl): Promise<void> {
    if (!command.id.trim()) throw new Error("Replay controls require a command ID.");
    if (
      command.type === "seek" &&
      (typeof command.offsetMs !== "number" || 
        !Number.isFinite(command.offsetMs) || 
        command.offsetMs < 0)
    ) {
      throw new Error("Replay seek offset must be a non-negative number.");
    }
    await invoke("control_replay", {
      command: {
        ...command,
        offsetMs: command.offsetMs === undefined ? undefined : Math.round(command.offsetMs),
      },
    });
  }

  async seekReplay(offsetMs: number): Promise<void> {
    await this.controlReplay({ id: `seek-${Date.now()}`, type: "seek", offsetMs });
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
