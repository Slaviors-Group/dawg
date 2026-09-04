import { useState } from "react";
import { CaptureControls } from "./components/CaptureControls";
import { Dashboard } from "./components/Dashboard";
import { DiffReport } from "./components/DiffReport";
import { PolicyConfig } from "./components/PolicyConfig";
import { ReplayViewer } from "./components/ReplayViewer";
import { EngineProvider, useEngine } from "./context/EngineContext";
import "./App.css";

type Tab = "dashboard" | "capture" | "replay" | "diff" | "policy";

function DoctorModal({ onClose }: { onClose: () => void }) {
  const { doctorReport, refetchEngineStatus, isCheckingEngine } = useEngine();

  return (
    <div className="fixed inset-0 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4 z-50">
      <div className="bg-slate-800 border border-slate-700 rounded-xl max-w-2xl w-full p-6 shadow-2xl flex flex-col gap-4">
        <div className="flex justify-between items-center border-b border-slate-700 pb-3">
          <div>
            <h3 className="text-lg font-bold text-slate-100 flex items-center gap-2">
              <span>🩺 DAWG Engine Doctor</span>
              {doctorReport && (
                <span
                  className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                    doctorReport.status === "ready"
                      ? "bg-emerald-500/20 text-emerald-300 border border-emerald-500/30"
                      : "bg-amber-500/20 text-amber-300 border border-amber-500/30"
                  }`}
                >
                  {doctorReport.status.toUpperCase()}
                </span>
              )}
            </h3>
            <p className="text-xs text-slate-400 mt-0.5">
              Runtime diagnostics and bundled resource verification
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="text-slate-400 hover:text-slate-100 p-1 text-lg font-bold"
          >
            ✕
          </button>
        </div>

        {doctorReport ? (
          <div className="flex flex-col gap-3 max-h-96 overflow-y-auto pr-1">
            <div className="bg-slate-900/60 p-3 rounded-lg border border-slate-700/50 text-xs flex flex-col gap-1">
              <div>
                <span className="text-slate-400">Engine Path: </span>
                <span className="text-sky-300 font-mono">{doctorReport.enginePath}</span>
              </div>
              <div>
                <span className="text-slate-400">Resource Root: </span>
                <span className="text-sky-300 font-mono">{doctorReport.resourceDir}</span>
                {doctorReport.isBundled && (
                  <span className="ml-2 bg-purple-500/20 text-purple-300 px-1.5 py-0.2 rounded text-[10px]">
                    Bundled App Mode
                  </span>
                )}
              </div>
            </div>

            <div className="flex flex-col gap-2">
              <h4 className="text-xs font-semibold text-slate-300">Runtime Components</h4>
              {doctorReport.components.map((c) => (
                <div
                  key={c.name}
                  className="flex items-center justify-between p-2.5 rounded-lg bg-slate-900/40 border border-slate-800 text-xs"
                >
                  <div className="flex items-center gap-2">
                    <span className="text-base">{c.installed ? "✅" : "❌"}</span>
                    <div>
                      <div className="font-medium text-slate-200 flex items-center gap-1.5">
                        <span>{c.name}</span>
                        {c.bundled && (
                          <span className="bg-sky-500/20 text-sky-300 px-1.5 py-0.2 rounded text-[10px]">
                            bundled
                          </span>
                        )}
                        {c.version && (
                          <span className="text-slate-400 font-mono text-[11px]">
                            v{c.version}
                          </span>
                        )}
                      </div>
                      {c.path && (
                        <div className="text-[10px] text-slate-500 font-mono truncate max-w-md">
                          {c.path}
                        </div>
                      )}
                      {c.error && (
                        <div className="text-[11px] text-rose-400 mt-0.5">{c.error}</div>
                      )}
                    </div>
                  </div>
                  <span
                    className={`px-2 py-0.5 rounded text-[10px] font-medium ${
                      c.installed
                        ? "bg-emerald-500/10 text-emerald-400"
                        : "bg-rose-500/10 text-rose-400"
                    }`}
                  >
                    {c.installed ? "Ready" : "Missing"}
                  </span>
                </div>
              ))}
            </div>
          </div>
        ) : (
          <div className="py-8 text-center text-slate-400 text-sm">
            {isCheckingEngine ? "Running diagnostics..." : "No diagnostic report available."}
          </div>
        )}

        <div className="flex justify-between items-center border-t border-slate-700 pt-3 mt-1">
          <button
            type="button"
            onClick={refetchEngineStatus}
            disabled={isCheckingEngine}
            className="px-3 py-1.5 rounded-md text-xs bg-slate-700 hover:bg-slate-600 text-slate-200 disabled:opacity-50 transition-colors"
          >
            {isCheckingEngine ? "Re-probing..." : "Re-run Doctor"}
          </button>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-1.5 rounded-md text-xs bg-sky-600 hover:bg-sky-500 text-white font-medium transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}

function ShellContent() {
  const [activeTab, setActiveTab] = useState<Tab>("dashboard");
  const [showDoctorModal, setShowDoctorModal] = useState(false);
  const { engineStatus, isCheckingEngine } = useEngine();

  return (
    <div className="flex flex-col h-screen bg-slate-900 text-slate-100 font-sans">
      <header className="flex justify-between items-center px-6 py-3 bg-slate-800 border-b border-slate-700">
        <div className="flex items-center gap-4">
          <div className="flex items-baseline gap-2">
            <span className="font-bold text-xl text-sky-400">DAWG</span>
            <span className="text-xs text-slate-400">Desktop Shell</span>
          </div>

          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setShowDoctorModal(true)}
              title="Click to view full Engine Doctor diagnostics"
              className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium border transition-colors cursor-pointer ${
                engineStatus.installed
                  ? "bg-emerald-500/10 border-emerald-500/30 text-emerald-300 hover:bg-emerald-500/20"
                  : "bg-amber-500/10 border-amber-500/30 text-amber-300 hover:bg-amber-500/20"
              }`}
            >
              <span
                className={`w-2 h-2 rounded-full ${
                  isCheckingEngine
                    ? "bg-slate-400 animate-pulse"
                    : engineStatus.installed
                      ? "bg-emerald-400"
                      : "bg-amber-400"
                }`}
              />
              <span>
                {isCheckingEngine
                  ? "Probing Engine..."
                  : engineStatus.installed
                    ? `Engine ${engineStatus.version || "Ready"}${engineStatus.bundled ? " [Bundled]" : ""}`
                    : "CLI Not Detected"}
              </span>
            </button>
          </div>
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

      {showDoctorModal && <DoctorModal onClose={() => setShowDoctorModal(false)} />}
    </div>
  );
}

function App() {
  return (
    <EngineProvider>
      <ShellContent />
    </EngineProvider>
  );
}

export default App;
