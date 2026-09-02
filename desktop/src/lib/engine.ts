import { invoke } from "@tauri-apps/api/core";

/**
 * Typed IPC bridge for DAWG Engine CLI subprocess invocations.
 * Implements IPC contract table from architecture.md §Desktop ↔ Engine IPC Contract.
 */

export interface StartCaptureOptions {
  url: string;
}

export interface StartCaptureResult {
  status: "capturing";
  sessionId: string;
}

export interface StopCaptureOptions {
  controlFile?: string;
}

export interface StopCaptureResult {
  status: "done";
  artifactPath: string;
}

export interface InspectOptions {
  path: string;
}

export interface InspectResult {
  id?: string;
  version?: string;
  [key: string]: unknown;
}

export interface RunReplayOptions {
  artifact: string;
}

export interface RunReplayResult {
  status: "pass" | "fail";
  report: Record<string, unknown>;
}

export interface VerifyOptions {
  artifact: string;
  against: string;
}

export interface VerifyResult {
  status: "pass" | "fail";
  diff: Record<string, unknown>;
}

export interface CommandOutput<T = unknown> {
  status: "success" | "error";
  payload: T;
}

export class EngineBridge {
  async startCapture(options: StartCaptureOptions): Promise<StartCaptureResult> {
    try {
      const res = await invoke<CommandOutput<StartCaptureResult>>("start_capture", {
        url: options.url,
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
