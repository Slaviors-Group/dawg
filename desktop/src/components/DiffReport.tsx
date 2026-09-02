import type React from "react";

export const DiffReport: React.FC = () => {
  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-6 max-w-4xl mx-auto">
      <h2 className="text-2xl font-semibold mb-1">Diff &amp; Verification Report</h2>
      <p className="text-slate-400 text-sm mb-6">
        Compare replayed outcome against target branch or expectations
      </p>

      <div className="mt-6 pt-4 border-t border-slate-700">
        <h3 className="text-lg font-medium mb-3">Verification Status</h3>
        <div className="bg-slate-900/30 border border-dashed border-slate-700 rounded-md p-8 text-center text-slate-400 text-sm">
          <p>
            No verification report generated yet. Run a verification task against a replay artifact.
          </p>
        </div>
      </div>
    </div>
  );
};
