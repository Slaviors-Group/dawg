import type React from "react";
import { useState } from "react";
import { useEngine } from "../context/EngineContext";
import { LogStreamer } from "./LogStreamer";

export const DiffReport: React.FC = () => {
  const { artifacts, addLogLine } = useEngine();
  const [selectedArtifact, setSelectedArtifact] = useState("");
  const [targetBranch, setTargetBranch] = useState("main");

  const handleRunVerify = () => {
    if (!selectedArtifact) return;
    addLogLine(`Running verify diff for artifact ${selectedArtifact} against branch ${targetBranch}...`);
  };

  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-6 max-w-4xl mx-auto space-y-6">
      <div>
        <h2 className="text-2xl font-semibold mb-1">Diff &amp; Verification Report</h2>
        <p className="text-slate-400 text-sm">
          Compare replayed outcome against target branch or baseline assertions
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="flex flex-col gap-1">
          <label htmlFor="verify-artifact-select" className="text-xs font-medium text-slate-300">
            Select Artifact
          </label>
          <select
            id="verify-artifact-select"
            value={selectedArtifact}
            onChange={(e) => setSelectedArtifact(e.target.value)}
            className="bg-slate-900 border border-slate-700 text-slate-100 px-3 py-2 rounded-md focus:outline-none focus:border-sky-500 text-sm"
          >
            <option value="">Select a captured .dawg artifact...</option>
            {artifacts.map((art) => (
              <option key={art.id} value={art.path}>
                {art.id} — {art.targetUrl}
              </option>
            ))}
          </select>
        </div>

        <div className="flex flex-col gap-1">
          <label htmlFor="target-branch" className="text-xs font-medium text-slate-300">
            Target Branch / Baseline
          </label>
          <div className="flex gap-2">
            <input
              id="target-branch"
              type="text"
              value={targetBranch}
              onChange={(e) => setTargetBranch(e.target.value)}
              className="flex-1 bg-slate-900 border border-slate-700 text-slate-100 px-3 py-2 rounded-md text-sm focus:outline-none focus:border-sky-500"
            />
            <button
              type="button"
              onClick={handleRunVerify}
              disabled={!selectedArtifact}
              className="px-4 py-2 bg-emerald-500 hover:bg-emerald-400 disabled:opacity-50 text-slate-950 font-semibold rounded-md transition-colors disabled:cursor-not-allowed text-sm"
            >
              Verify
            </button>
          </div>
        </div>
      </div>

      <div className="pt-4 border-t border-slate-700">
        <h3 className="text-lg font-medium mb-3">Verification Report</h3>
        <div className="bg-slate-900/30 border border-dashed border-slate-700 rounded-md p-8 text-center text-slate-400 text-sm">
          <p>No verification report generated yet. Select an artifact and run verification.</p>
        </div>
      </div>

      <div className="pt-4 border-t border-slate-700">
        <LogStreamer />
      </div>
    </div>
  );
};
