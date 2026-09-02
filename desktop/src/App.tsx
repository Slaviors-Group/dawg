import { useState } from "react";
import { CaptureControls } from "./components/CaptureControls";
import { Dashboard } from "./components/Dashboard";
import { DiffReport } from "./components/DiffReport";
import { PolicyConfig } from "./components/PolicyConfig";
import { ReplayViewer } from "./components/ReplayViewer";
import "./App.css";

type Tab = "dashboard" | "capture" | "replay" | "diff" | "policy";

function App() {
  const [activeTab, setActiveTab] = useState<Tab>("dashboard");

  return (
    <div className="layout">
      <header className="navbar">
        <div className="brand">
          <span className="brand-name">DAWG</span>
          <span className="brand-subtitle">Desktop Shell</span>
        </div>
        <nav className="nav-tabs">
          <button
            type="button"
            className={`tab-btn ${activeTab === "dashboard" ? "active" : ""}`}
            onClick={() => setActiveTab("dashboard")}
          >
            Dashboard
          </button>
          <button
            type="button"
            className={`tab-btn ${activeTab === "capture" ? "active" : ""}`}
            onClick={() => setActiveTab("capture")}
          >
            Capture
          </button>
          <button
            type="button"
            className={`tab-btn ${activeTab === "replay" ? "active" : ""}`}
            onClick={() => setActiveTab("replay")}
          >
            Replay
          </button>
          <button
            type="button"
            className={`tab-btn ${activeTab === "diff" ? "active" : ""}`}
            onClick={() => setActiveTab("diff")}
          >
            Diff &amp; Verify
          </button>
          <button
            type="button"
            className={`tab-btn ${activeTab === "policy" ? "active" : ""}`}
            onClick={() => setActiveTab("policy")}
          >
            Sanitizer Policy
          </button>
        </nav>
      </header>

      <main className="content">
        {activeTab === "dashboard" && <Dashboard />}
        {activeTab === "capture" && <CaptureControls />}
        {activeTab === "replay" && <ReplayViewer />}
        {activeTab === "diff" && <DiffReport />}
        {activeTab === "policy" && <PolicyConfig />}
      </main>
    </div>
  );
}

export default App;
