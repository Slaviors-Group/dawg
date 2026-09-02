import type React from "react";

export const ReplayViewer: React.FC = () => {
  return (
    <div className="panel">
      <h2>Replay Engine Viewer</h2>
      <p className="subtitle">Execute deterministic sandboxed artifact replay</p>

      <div className="form-group">
        <label htmlFor="artifact-select">Select Artifact</label>
        <div className="input-row">
          <select id="artifact-select" disabled>
            <option value="">No artifacts available</option>
          </select>
          <button type="button" className="btn btn-primary" disabled>
            Run Replay
          </button>
        </div>
      </div>

      <div className="section">
        <h3>Replay Sandbox Environment</h3>
        <div className="empty-state">
          <p>
            Requires Linux / WSL2 rootless Docker sandbox. Select an artifact to view replay status.
          </p>
        </div>
      </div>
    </div>
  );
};
