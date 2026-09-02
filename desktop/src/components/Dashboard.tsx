import type React from "react";

export const Dashboard: React.FC = () => {
  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-6 max-w-4xl mx-auto">
      <h2 className="text-2xl font-semibold mb-1">DAWG Dashboard</h2>
      <p className="text-slate-400 text-sm mb-6">System Status &amp; Recent Capture Artifacts</p>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <div className="bg-slate-900/50 border border-slate-700 rounded-md p-4 flex flex-col gap-1">
          <span className="text-xs text-slate-400 uppercase tracking-wider">Engine Status</span>
          <span className="text-xl font-semibold text-emerald-400">Ready</span>
        </div>
        <div className="bg-slate-900/50 border border-slate-700 rounded-md p-4 flex flex-col gap-1">
          <span className="text-xs text-slate-400 uppercase tracking-wider">Total Artifacts</span>
          <span className="text-xl font-semibold text-slate-100">0</span>
        </div>
        <div className="bg-slate-900/50 border border-slate-700 rounded-md p-4 flex flex-col gap-1">
          <span className="text-xs text-slate-400 uppercase tracking-wider">Sanitizer Policy</span>
          <span className="text-xl font-semibold text-slate-100">default.rego</span>
        </div>
      </div>

      <div className="mt-6 pt-4 border-t border-slate-700">
        <h3 className="text-lg font-medium mb-3">Recent Artifacts</h3>
        <div className="bg-slate-900/30 border border-dashed border-slate-700 rounded-md p-8 text-center text-slate-400 text-sm">
          <p>No captured artifacts found. Start a capture session to generate .dawg artifacts.</p>
        </div>
      </div>
    </div>
  );
};
