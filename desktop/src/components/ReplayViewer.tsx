import type React from "react";
import { useState } from "react";
import { useEngine } from "../context/EngineContext";
import { LogStreamer } from "./LogStreamer";

export const ReplayViewer: React.FC = () => {
  const { artifacts, addLogLine } = useEngine();
  const [selectedArtifact, setSelectedArtifact] = useState("");

  const handleRunReplay = () => {
    if (!selectedArtifact) return;
    addLogLine(`Triggering sandboxed replay for artifact: ${selectedArtifact}`);
  };

  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-6 max-w-4xl mx-auto space-y-6">
      <div>
        <h2 className="text-2xl font-semibold mb-1">Replay Engine Viewer</h2>
        <p className="text-slate-400 text-sm">Execute deterministic sandboxed artifact replay</p>
      </div>

      <div className="flex flex-col gap-2">
        <label htmlFor="artifact-select" className="text-sm font-medium text-slate-300">
          Select Artifact
        </label>
        <div className="flex gap-2">
          <select
            id="artifact-select"
            value={selectedArtifact}
            onChange={(e) => setSelectedArtifact(e.target.value)}
            className="flex-1 bg-slate-900 border border-slate-700 text-slate-100 px-3 py-2 rounded-md focus:outline-none focus:border-sky-500"
          >
            <option value="">Select a captured .dawg artifact...</option>
            {artifacts.map((art) => (
              <option key={art.id} value={art.path}>
                {art.id} — {art.targetUrl} ({new Date(art.createdAt).toLocaleTimeString()})
              </option>
            ))}
          </select>
          <button
            type="button"
            onClick={handleRunReplay}
            disabled={!selectedArtifact}
            className="px-4 py-2 bg-sky-500 hover:bg-sky-400 disabled:opacity-50 text-slate-950 font-semibold rounded-md transition-colors disabled:cursor-not-allowed"
          >
            Run Replay
          </button>
        </div>
      </div>

      <div className="pt-4 border-t border-slate-700">
        <h3 className="text-lg font-medium mb-3">Replay Sandbox Environment</h3>
        <div className="bg-slate-900/30 border border-dashed border-slate-700 rounded-md p-6 text-center text-slate-400 text-sm">
          <p className="mb-1">Requires Linux / WSL2 rootless Docker sandbox.</p>
          <p className="text-xs text-slate-500">
            Bare Windows hosts return sandbox isolation requirement error per security architecture.
          </p>
        </div>
      </div>

      <div className="pt-4 border-t border-slate-700">
        <LogStreamer />
      </div>
    </div>
  );
};
