import type React from "react";

export const Dashboard: React.FC = () => {
  return (
    <div className="panel">
      <h2>DAWG Dashboard</h2>
      <p className="subtitle">System Status &amp; Recent Capture Artifacts</p>

      <div className="card-grid">
        <div className="card">
          <span className="card-label">Engine Status</span>
          <span className="card-value status-idle">Ready</span>
        </div>
        <div className="card">
          <span className="card-label">Total Artifacts</span>
          <span className="card-value">0</span>
        </div>
        <div className="card">
          <span className="card-label">Sanitizer Policy</span>
          <span className="card-value">default.rego</span>
        </div>
      </div>

      <div className="section">
        <h3>Recent Artifacts</h3>
        <div className="empty-state">
          <p>No captured artifacts found. Start a capture session to generate .dawg artifacts.</p>
        </div>
      </div>
    </div>
  );
};
