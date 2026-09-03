import type React from "react";
import { useEngine } from "../context/EngineContext";

export const EngineHelpBanner: React.FC = () => {
  const { engineStatus, refetchEngineStatus, isCheckingEngine } = useEngine();

  if (engineStatus.installed) {
    return null;
  }

  return (
    <div className="bg-amber-500/10 border border-amber-500/30 rounded-lg p-5 text-amber-200 text-sm space-y-3 mb-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 font-semibold text-amber-300">
          <svg
            className="w-5 h-5 text-amber-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            aria-label="Warning"
          >
            <title>Warning</title>
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
            />
          </svg>
          <span>DAWG Engine CLI Binary Not Found on PATH</span>
        </div>
        <button
          type="button"
          onClick={refetchEngineStatus}
          disabled={isCheckingEngine}
          className="px-3 py-1.5 bg-amber-500/20 hover:bg-amber-500/30 border border-amber-500/40 rounded text-xs font-semibold text-amber-200 transition-colors disabled:opacity-50"
        >
          {isCheckingEngine ? "Probing..." : "Re-probe PATH"}
        </button>
      </div>

      <p className="text-xs text-amber-200/80 leading-relaxed">
        The Desktop Shell uses Tauri IPC to spawn `dawg` CLI subprocesses. To build and install the
        engine binary locally:
      </p>

      <div className="bg-slate-950 border border-slate-800 p-3 rounded font-mono text-xs text-slate-200 overflow-x-auto space-y-1">
        <div>
          <span className="text-slate-500"># 1. Navigate to engine directory</span>
        </div>
        <div>
          <span className="text-sky-400">cd</span> dawg/engine
        </div>
        <div>
          <span className="text-slate-500"># 2. Build the CLI binary</span>
        </div>
        <div>
          <span className="text-sky-400">go build</span> -o dawg ./cmd/dawg
        </div>
        <div>
          <span className="text-slate-500"># 3. Add to system PATH or execute dawg init</span>
        </div>
      </div>
    </div>
  );
};
