import type React from "react";

export const ReplayViewer: React.FC = () => {
  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-6 max-w-4xl mx-auto">
      <h2 className="text-2xl font-semibold mb-1">Replay Engine Viewer</h2>
      <p className="text-slate-400 text-sm mb-6">Execute deterministic sandboxed artifact replay</p>

      <div className="flex flex-col gap-2 mb-6">
        <label htmlFor="artifact-select" className="text-sm font-medium text-slate-300">
          Select Artifact
        </label>
        <div className="flex gap-2">
          <select
            id="artifact-select"
            disabled
            className="flex-1 bg-slate-900 border border-slate-700 text-slate-100 px-3 py-2 rounded-md disabled:opacity-60"
          >
            <option value="">No artifacts available</option>
          </select>
          <button
            type="button"
            disabled
            className="px-4 py-2 bg-sky-500 text-slate-950 font-semibold rounded-md disabled:opacity-50 cursor-not-allowed"
          >
            Run Replay
          </button>
        </div>
      </div>

      <div className="mt-6 pt-4 border-t border-slate-700">
        <h3 className="text-lg font-medium mb-3">Replay Sandbox Environment</h3>
        <div className="bg-slate-900/30 border border-dashed border-slate-700 rounded-md p-8 text-center text-slate-400 text-sm">
          <p>
            Requires Linux / WSL2 rootless Docker sandbox. Select an artifact to view replay status.
          </p>
        </div>
      </div>
    </div>
  );
};
