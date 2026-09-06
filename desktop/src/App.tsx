import { useState } from "react";
import { AnimatePresence, motion } from "framer-motion";
import {
  SquaresFour,
  Record,
  ArrowCounterClockwise,
  GitDiff,
  ShieldCheck,
  GearSix,
  GithubLogo,
} from "@phosphor-icons/react";

import { CaptureControls } from "./components/CaptureControls";
import { Dashboard } from "./components/Dashboard";
import { DiffReport } from "./components/DiffReport";
import { PolicyConfig } from "./components/PolicyConfig";
import { ReplayViewer } from "./components/ReplayViewer";
import { NavItem } from "./components/ui/NavItem";
import { SettingsPanel } from "./components/ui/SettingsPanel";

import { EngineProvider, useEngine } from "./context/EngineContext";
import { ThemeProvider } from "./context/ThemeContext";
import { MotionProvider } from "./context/MotionContext";
import { ToastProvider } from "./components/ui/Toast";

import { DoctorModal } from "./components/DoctorModal";
import "./App.css";

type Tab = "dashboard" | "capture" | "replay" | "diff" | "policy";

const NAV_ITEMS: {
  id: Tab;
  label: string;
  icon: typeof SquaresFour;
}[] = [
  { id: "dashboard", label: "Dashboard",       icon: SquaresFour },
  { id: "capture",   label: "Capture",          icon: Record },
  { id: "replay",    label: "Replay",           icon: ArrowCounterClockwise },
  { id: "diff",      label: "Diff & Verify",    icon: GitDiff },
  { id: "policy",    label: "Sanitizer Policy", icon: ShieldCheck },
];

function ShellContent() {
  const [activeTab, setActiveTab] = useState<Tab>("dashboard");
  const [showDoctor, setShowDoctor] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const { engineStatus, isCheckingEngine, isCapturing } = useEngine();

  return (
    <div className="flex h-screen bg-canvas overflow-hidden">
      
      {/* ─── Sidebar ─── */}
      <aside className="w-[260px] shrink-0 flex flex-col bg-surface border-r border-border py-4 px-4 gap-6 z-20">

        {/* App Logo / Top header */}
        <div className="flex items-center gap-3 p-2 mb-2">
          <img src="/paw-dawg.svg" alt="DAWG Logo" className="w-10 h-10 shrink-0" />
          <div>
            <span className="font-bold text-[15px] text-text-primary block leading-none mb-1 tracking-tight">
              DAWG
            </span>
            <span className="text-xs text-text-tertiary font-medium">Digs Any Web-app Glitch</span>
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex flex-col gap-1 flex-1" aria-label="Main navigation">
          {NAV_ITEMS.map(({ id, label, icon: Icon }) => (
            <NavItem
              key={id}
              icon={<Icon size={18} weight={activeTab === id ? "duotone" : "regular"} />}
              label={label}
              active={activeTab === id}
              onClick={() => setActiveTab(id)}
              badge={
                id === "capture" && isCapturing ? (
                  <span className="w-2 h-2 rounded-full bg-error-dot animate-pulse shrink-0" />
                ) : undefined
              }
            />
          ))}

          {/* Engine status pill (replacing the old style) */}
          <div className="mt-8 pt-6 border-t border-border">
            <span className="text-[10px] font-bold text-text-tertiary uppercase tracking-wider block mb-2 px-3">
              Engine Status
            </span>
            <button
              type="button"
              onClick={() => setShowDoctor(true)}
              className="w-full flex items-center gap-3 px-3 py-2 rounded-md hover:bg-surface-hover transition-colors duration-fast"
            >
              <span
                className={[
                  "w-2 h-2 rounded-full shrink-0",
                  isCheckingEngine
                    ? "bg-text-tertiary animate-pulse"
                    : engineStatus.installed
                      ? "bg-success-dot"
                      : "bg-warning-dot",
                ].join(" ")}
              />
              <span className="text-xs font-medium text-text-secondary truncate flex-1 text-left">
                {isCheckingEngine
                  ? "Probing…"
                  : engineStatus.installed
                    ? `Ready (v${engineStatus.version ?? "latest"})`
                    : "CLI Not Detected"}
              </span>
            </button>
            <button
              type="button"
              onClick={() => setShowSettings(true)}
              className="w-full flex items-center gap-3 px-3 py-2 rounded-md hover:bg-surface-hover transition-colors duration-fast text-text-secondary"
            >
              <GearSix size={18} className="text-text-tertiary" />
              <span className="text-xs font-medium">Settings</span>
            </button>
          </div>
        </nav>

        {/* Open Source / GitHub Link */}
        <div className="relative p-5 rounded-2xl overflow-hidden shadow-card shrink-0">
          {/* Gradient Background */}
          <div className="absolute inset-0 bg-gradient-to-br from-brand-600 to-brand-400" />
          
          {/* Dot Pattern Overlay */}
          <div 
            className="absolute inset-0 opacity-30" 
            style={{ 
              backgroundImage: 'radial-gradient(circle at 1px 1px, white 1px, transparent 0)',
              backgroundSize: '12px 12px' 
            }} 
          />
          
          {/* Glass glare effect */}
          <div className="absolute top-0 right-0 -mr-8 -mt-8 w-32 h-32 bg-white opacity-20 rounded-full blur-2xl mix-blend-overlay" />
          
          <div className="relative z-10 flex flex-col items-start gap-3">
            <div className="flex items-center gap-2">
              <GithubLogo size={18} weight="fill" className="text-white" />
              <span className="font-bold text-sm text-white">Open Source</span>
            </div>
            <p className="text-xs text-brand-50 leading-relaxed font-medium">
              DAWG is completely free and open-source. Help us improve!
            </p>
            <a
              href="https://github.com/Slaviors-Group/dawg"
              target="_blank"
              rel="noopener noreferrer"
              className="mt-1 px-4 py-1.5 bg-white text-brand-700 hover:bg-brand-50 font-semibold text-xs rounded-full shadow-sm transition-colors w-full text-center block"
            >
              Star on GitHub
            </a>
          </div>
        </div>
      </aside>

      {/* ─── Main content ─── */}
      <main className="flex-1 relative min-w-0 overflow-y-auto bg-grid-pattern z-10">
        
        {/* Glow Orbs in background */}
        <div className="glow-orb top-[-100px] left-[-100px]" />
        <div className="glow-orb-cyan right-[-50px] top-[150px]" />
        
        <div className="relative px-10 pt-10 min-h-full flex flex-col">
          <AnimatePresence mode="wait">
            <motion.div
              key={activeTab}
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ duration: 0.2, ease: "easeOut" }}
              className="min-h-full flex flex-col"
            >
              {activeTab === "dashboard" && <Dashboard />}
              {activeTab === "capture"   && <CaptureControls />}
              {activeTab === "replay"    && <ReplayViewer />}
              {activeTab === "diff"      && <DiffReport />}
              {activeTab === "policy"    && <PolicyConfig />}
            </motion.div>
          </AnimatePresence>
          {/* Explicit spacer to force scroll padding at the bottom */}
          <div className="h-10 shrink-0 w-full" />
        </div>
      </main>

      {/* Modals */}
      <DoctorModal open={showDoctor} onClose={() => setShowDoctor(false)} />
      <SettingsPanel open={showSettings} onClose={() => setShowSettings(false)} />
    </div>
  );
}

function App() {
  return (
    <ThemeProvider>
      <MotionProvider>
        <ToastProvider>
          <EngineProvider>
            <ShellContent />
          </EngineProvider>
        </ToastProvider>
      </MotionProvider>
    </ThemeProvider>
  );
}

export default App;
