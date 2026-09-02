/**
 * Typed IPC bridge for DAWG Engine CLI subprocess invocations.
 * Implements IPC contract table from architecture.md §Desktop ↔ Engine IPC Contract.
 *
 * Note: Subprocess execution is explicitly deferred (T0.2 / T2.1 foundation phase).
 * All methods throw "not implemented" error until CLI subprocess invocation is wired.
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

export class EngineBridge {
  async startCapture(_options: StartCaptureOptions): Promise<StartCaptureResult> {
    throw new Error("EngineBridge.startCapture is not implemented");
  }

  async stopCapture(_options?: StopCaptureOptions): Promise<StopCaptureResult> {
    throw new Error("EngineBridge.stopCapture is not implemented");
  }

  async inspectArtifact(_options: InspectOptions): Promise<InspectResult> {
    throw new Error("EngineBridge.inspectArtifact is not implemented");
  }

  async runReplay(_options: RunReplayOptions): Promise<RunReplayResult> {
    throw new Error("EngineBridge.runReplay is not implemented");
  }

  async verifyResult(_options: VerifyOptions): Promise<VerifyResult> {
    throw new Error("EngineBridge.verifyResult is not implemented");
  }
}

export const engine = new EngineBridge();
