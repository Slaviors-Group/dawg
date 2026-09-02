import type React from "react";

export const CaptureControls: React.FC = () => {
  return (
    <div className="panel">
      <h2>Capture Session Controls</h2>
      <p className="subtitle">Record browser, HTTP traffic, database diffs, and structured logs</p>

      <form onSubmit={(e) => e.preventDefault()} className="form-group">
        <label htmlFor="target-url">Target Application URL</label>
        <div className="input-row">
          <input
            id="target-url"
            type="url"
            placeholder="https://staging.example.com"
            defaultValue="http://localhost:3000"
            disabled
          />
          <button type="button" className="btn btn-primary" disabled>
            Start Capture
          </button>
          <button type="button" className="btn btn-secondary" disabled>
            Stop Capture
          </button>
        </div>
      </form>

      <div className="section">
        <h3>Session Configuration</h3>
        <ul className="config-list">
          <li>
            <strong>Browser Capture:</strong> Playwright + rrweb trace enabled
          </li>
          <li>
            <strong>HTTP Proxy:</strong> mitmproxy cassette capture enabled
          </li>
          <li>
            <strong>Database Tap:</strong> ORM diff collector enabled
          </li>
          <li>
            <strong>Log Capture:</strong> Application stdout / JSONL stream enabled
          </li>
        </ul>
      </div>
    </div>
  );
};
