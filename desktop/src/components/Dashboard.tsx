import type React from "react";
import { useEngine } from "../context/EngineContext";
import { ArtifactInspectorModal } from "./ArtifactInspectorModal";
import { ArtifactList } from "./ArtifactList";
import { EngineHelpBanner } from "./EngineHelpBanner";
import { LogStreamer } from "./LogStreamer";

export const Dashboard: React.FC = () => {
  const { engineStatus, artifacts, inspectedArtifact, closeInspectModal } = useEngine();

  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-6 max-w-4xl mx-auto space-y-6">
      <div>
        <h2 className="text-2xl font-semibold mb-1">DAWG Dashboard</h2>
        <p className="text-slate-400 text-sm">System Status &amp; Recent Capture Artifacts</p>
      </div>

      <EngineHelpBanner />

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-slate-900/50 border border-slate-700 rounded-md p-4 flex flex-col gap-1">
          <span className="text-xs text-slate-400 uppercase tracking-wider">Engine Status</span>
          <span
            className={`text-xl font-semibold ${
              engineStatus.installed ? "text-emerald-400" : "text-amber-400"
            }`}
          >
            {engineStatus.installed ? "Ready" : "CLI Not Detected"}
          </span>
        </div>
        <div className="bg-slate-900/50 border border-slate-700 rounded-md p-4 flex flex-col gap-1">
          <span className="text-xs text-slate-400 uppercase tracking-wider">Total Artifacts</span>
          <span className="text-xl font-semibold text-slate-100">{artifacts.length}</span>
        </div>
        <div className="bg-slate-900/50 border border-slate-700 rounded-md p-4 flex flex-col gap-1">
          <span className="text-xs text-slate-400 uppercase tracking-wider">Sanitizer Policy</span>
          <span className="text-xl font-semibold text-slate-100">default.rego</span>
        </div>
      </div>

      <div className="pt-4 border-t border-slate-700 space-y-3">
        <h3 className="text-lg font-medium">Recent Artifacts</h3>
        <ArtifactList />
      </div>

      <div className="pt-4 border-t border-slate-700">
        <LogStreamer />
      </div>

      <ArtifactInspectorModal artifact={inspectedArtifact} onClose={closeInspectModal} />
    </div>
  );
};
