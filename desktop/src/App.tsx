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
    <div className="flex flex-col h-screen bg-slate-900 text-slate-100 font-sans">
      <header className="flex justify-between items-center px-6 py-3 bg-slate-800 border-b border-slate-700">
        <div className="flex items-baseline gap-2">
          <span className="font-bold text-xl text-sky-400">DAWG</span>
          <span className="text-xs text-slate-400">Desktop Shell</span>
        </div>
        <nav className="flex gap-2">
          <button
            type="button"
            className={`px-4 py-2 rounded-md text-sm transition-all border ${
              activeTab === "dashboard"
                ? "bg-slate-700 border-sky-400 text-slate-100"
                : "bg-transparent border-transparent text-slate-400 hover:text-slate-100 hover:bg-slate-800/50"
            }`}
            onClick={() => setActiveTab("dashboard")}
          >
            Dashboard
          </button>
          <button
            type="button"
            className={`px-4 py-2 rounded-md text-sm transition-all border ${
              activeTab === "capture"
                ? "bg-slate-700 border-sky-400 text-slate-100"
                : "bg-transparent border-transparent text-slate-400 hover:text-slate-100 hover:bg-slate-800/50"
            }`}
            onClick={() => setActiveTab("capture")}
          >
            Capture
          </button>
          <button
            type="button"
            className={`px-4 py-2 rounded-md text-sm transition-all border ${
              activeTab === "replay"
                ? "bg-slate-700 border-sky-400 text-slate-100"
                : "bg-transparent border-transparent text-slate-400 hover:text-slate-100 hover:bg-slate-800/50"
            }`}
            onClick={() => setActiveTab("replay")}
          >
            Replay
          </button>
          <button
            type="button"
            className={`px-4 py-2 rounded-md text-sm transition-all border ${
              activeTab === "diff"
                ? "bg-slate-700 border-sky-400 text-slate-100"
                : "bg-transparent border-transparent text-slate-400 hover:text-slate-100 hover:bg-slate-800/50"
            }`}
            onClick={() => setActiveTab("diff")}
          >
            Diff &amp; Verify
          </button>
          <button
            type="button"
            className={`px-4 py-2 rounded-md text-sm transition-all border ${
              activeTab === "policy"
                ? "bg-slate-700 border-sky-400 text-slate-100"
                : "bg-transparent border-transparent text-slate-400 hover:text-slate-100 hover:bg-slate-800/50"
            }`}
            onClick={() => setActiveTab("policy")}
          >
            Sanitizer Policy
          </button>
        </nav>
      </header>

      <main className="flex-1 p-6 overflow-y-auto">
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
