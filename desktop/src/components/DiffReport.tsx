import type React from "react";

export const DiffReport: React.FC = () => {
  return (
    <div className="panel">
      <h2>Diff &amp; Verification Report</h2>
      <p className="subtitle">Compare replayed outcome against target branch or expectations</p>

      <div className="section">
        <h3>Verification Status</h3>
        <div className="empty-state">
          <p>
            No verification report generated yet. Run a verification task against a replay artifact.
          </p>
        </div>
      </div>
    </div>
  );
};
