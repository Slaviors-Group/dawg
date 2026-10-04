import {
  Archive,
  ArrowCounterClockwise,
  Browser,
  GearSix,
  GithubLogo,
  Heart,
  PencilSimple,
  Record,
  ShieldCheck,
  SquaresFour,
} from "@phosphor-icons/react";
import { invoke } from "@tauri-apps/api/core";
import { openUrl } from "@tauri-apps/plugin-opener";
import { AnimatePresence, motion } from "framer-motion";
import { useCallback, useEffect, useRef, useState } from "react";

import { ArtifactsPage } from "./components/ArtifactsPage";
import { CaptureControls } from "./components/CaptureControls";
import { Dashboard } from "./components/Dashboard";
import { EditorPage } from "./components/EditorPage";
import { PolicyConfig } from "./components/PolicyConfig";
import { ReplayViewer } from "./components/ReplayViewer";
import { NavItem } from "./components/ui/NavItem";
import { SettingsPanel } from "./components/ui/SettingsPanel";

import { ToastProvider, useToast } from "./components/ui/Toast";
import { type ArtifactItem, EngineProvider, useEngine } from "./context/EngineContext";
import { MotionProvider } from "./context/MotionContext";
import { ThemeProvider } from "./context/ThemeContext";

import { DoctorModal } from "./components/DoctorModal";
import { GuidedTour } from "./components/GuidedTour";
import { OnboardingPage } from "./components/OnboardingPage";
import "./App.css";

type Tab = "dashboard" | "capture" | "artifacts" | "replay" | "editor" | "policy";

const CHROME_WEB_STORE_EXTENSION_URL =
  "https://chromewebstore.google.com/detail/peiigoeakholhhbbbbfkeojomekmmokj?utm_source=item-share-cb";
const GITHUB_SPONSORS_URL = "https://github.com/sponsors/Slaviors-Group";
const GITHUB_GROUP_URL = "https://github.com/Slaviors-Group";
const BUY_ME_A_COFFEE_URL = "https://buymeacoffee.com/slaviorsgroup";
const ONBOARDING_COMPLETE_KEY = "dawg-onboarding-complete";
const GUIDE_COMPLETE_KEY = "dawg-workspace-guide-complete";

function hasCompletedOnboarding() {
  try {
    return localStorage.getItem(ONBOARDING_COMPLETE_KEY) === "true";
  } catch {
    return false;
  }
}

function hasCompletedGuide() {
  try {
    return localStorage.getItem(GUIDE_COMPLETE_KEY) === "true";
  } catch {
    return false;
  }
}

const NAV_ITEMS: {
  id: Tab;
  label: string;
  icon: typeof SquaresFour;
}[] = [
  { id: "dashboard", label: "Dashboard", icon: SquaresFour },
  { id: "capture", label: "Capture", icon: Record },
  { id: "artifacts", label: "Artifacts", icon: Archive },
  { id: "replay", label: "Replay", icon: ArrowCounterClockwise },
  { id: "editor", label: "Editor", icon: PencilSimple },
  { id: "policy", label: "Sanitizer Policy", icon: ShieldCheck },
];

function ShellContent() {
  const [activeTab, setActiveTab] = useState<Tab>("dashboard");
  const [showOnboarding, setShowOnboarding] = useState(() => !hasCompletedOnboarding());
  const [showGuide, setShowGuide] = useState(
    () => hasCompletedOnboarding() && !hasCompletedGuide(),
  );
  const [guideStep, setGuideStep] = useState(0);
  const [replayArtifactIdentity, setReplayArtifactIdentity] = useState<string | undefined>();
  const [editorArtifactIdentity, setEditorArtifactIdentity] = useState<string | undefined>();
  const [showDoctor, setShowDoctor] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const startupArtifactHandled = useRef(false);
  const { engineStatus, isCheckingEngine, isCapturing, addLogLine, importArtifact } = useEngine();
  const { error: showToastError } = useToast();

  useEffect(() => {
    if (startupArtifactHandled.current) return;
    startupArtifactHandled.current = true;

    void invoke<string | null>("startup_artifact")
      .then(async (archive) => {
        if (!archive) return;
        // An artifact opened through file association should remain visible on launch.
        // The welcome page will still appear on the next ordinary launch.
        setShowOnboarding(false);
        setActiveTab("dashboard");
        addLogLine(`Opening DAWG artifact ${archive}.`);
        try {
          await importArtifact(archive);
        } catch {
          // EngineContext records the actionable error in the shared log stream.
        }
      })
      .catch((error) => addLogLine(`[WARN] Could not open the launch artifact: ${String(error)}`));
  }, [addLogLine, importArtifact]);

  const artifactIdentity = (artifact: ArtifactItem) => artifact.instanceId || artifact.path;

  const handleReplayArtifact = (artifact: ArtifactItem) => {
    setReplayArtifactIdentity(artifactIdentity(artifact));
    setActiveTab("replay");
  };

  const handleEditArtifact = (artifact: ArtifactItem) => {
    setEditorArtifactIdentity(artifactIdentity(artifact));
    setActiveTab("editor");
  };

  const handleSponsor = async () => {
    try {
      await openUrl(GITHUB_SPONSORS_URL);
    } catch (error) {
      addLogLine(`[ERROR] Could not open the DAWG Sponsors page: ${String(error)}`);
    }
  };

  const handleInstallWebExtension = async () => {
    try {
      await openUrl(CHROME_WEB_STORE_EXTENSION_URL);
    } catch (error) {
      addLogLine(`[ERROR] Could not open the DAWG Web Extension page: ${String(error)}`);
      showToastError("Could not open the browser extension page", String(error));
    }
  };

  const openOnboardingLink = async (url: string, label: string) => {
    try {
      await openUrl(url);
    } catch (error) {
      addLogLine(`[ERROR] Could not open ${label}: ${String(error)}`);
      showToastError(`Could not open ${label}`, String(error));
    }
  };

  const finishOnboarding = () => {
    try {
      localStorage.setItem(ONBOARDING_COMPLETE_KEY, "true");
    } catch {
      // Continue into the workspace even if the WebView cannot persist preferences.
    }
    setActiveTab("dashboard");
    setGuideStep(0);
    setShowGuide(true);
    setShowOnboarding(false);
  };

  const finishGuide = useCallback(() => {
    try {
      localStorage.setItem(GUIDE_COMPLETE_KEY, "true");
    } catch {
      // The guide can still close if preferences cannot be persisted.
    }
    setShowGuide(false);
  }, []);

  if (showOnboarding) {
    return (
      <OnboardingPage
        onExplore={finishOnboarding}
        onInstallExtension={() => void handleInstallWebExtension()}
        onOpenGitHub={() => void openOnboardingLink(GITHUB_GROUP_URL, "Slaviors Group")}
        onOpenCoffee={() => void openOnboardingLink(BUY_ME_A_COFFEE_URL, "Buy Me a Coffee")}
        onDonate={() => void openOnboardingLink(GITHUB_SPONSORS_URL, "GitHub Sponsors")}
        links={{
          extension: CHROME_WEB_STORE_EXTENSION_URL,
          github: GITHUB_GROUP_URL,
          coffee: BUY_ME_A_COFFEE_URL,
          donate: GITHUB_SPONSORS_URL,
        }}
      />
    );
  }

  return (
    <div className="flex h-screen overflow-hidden bg-canvas">
      <aside className="z-20 flex h-screen w-65 shrink-0 flex-col gap-6 overflow-hidden border-r border-border/70 bg-surface px-4 py-4">
        <div className="flex items-center gap-3 p-2 mb-2">
          <img src="/paw-dawg.svg" alt="DAWG Logo" className="w-10 h-10 shrink-0" />
          <div>
            <span className="font-bold text-[15px] text-text-primary block leading-none mb-1 tracking-tight">
              DAWG
            </span>
            <span className="text-xs text-text-tertiary font-medium">Digs Any Web-app Glitch</span>
          </div>
        </div>

        <nav
          className="flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto overscroll-contain pr-1"
          aria-label="Main navigation"
        >
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

          <div className="mt-8 pt-6 border-t border-border">
            <span className="text-[10px] font-bold text-text-tertiary uppercase tracking-wider block mb-2 px-3">
              Engine Status
            </span>
            <button
              type="button"
              onClick={() => setShowDoctor(true)}
              className="w-full flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-surface-hover transition-colors duration-fast"
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
              className="w-full flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-surface-hover transition-colors duration-fast text-text-secondary"
            >
              <GearSix size={18} className="text-text-tertiary" />
              <span className="text-xs font-medium">Settings</span>
            </button>
            <button
              type="button"
              onClick={() => void handleInstallWebExtension()}
              className="w-full flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-surface-hover transition-colors duration-fast text-text-secondary"
            >
              <Browser size={18} className="text-text-tertiary" />
              <span className="text-xs font-medium">Install DAWG Web Extension</span>
            </button>
          </div>
        </nav>

        <div className="relative shrink-0 overflow-hidden rounded-3xl p-5 shadow-card">
          <div className="absolute inset-0 bg-linear-to-br from-[hsl(258_65%_45%)] to-[hsl(258_65%_65%)]" />
          <div
            className="absolute inset-0 opacity-30"
            style={{
              backgroundImage: "radial-gradient(circle at 1px 1px, white 1px, transparent 0)",
              backgroundSize: "12px 12px",
            }}
          />
          <div className="absolute top-0 right-0 -mr-8 -mt-8 w-32 h-32 rounded-full bg-white opacity-20 blur-2xl mix-blend-overlay" />
          <div className="relative z-10 flex flex-col items-start gap-3">
            <div className="flex items-center gap-2">
              <GithubLogo size={18} weight="fill" className="text-white" />
              <span className="font-bold text-sm text-white">Open Source</span>
            </div>
            <p className="text-xs font-medium leading-relaxed text-[hsl(258_75%_98%)]">
              Source code, releases, and issue tracking.
            </p>
            <a
              href="https://github.com/Slaviors-Group/dawg"
              target="_blank"
              rel="noopener noreferrer"
              className="mt-1 block w-full rounded-full bg-white px-4 py-1.5 text-center text-xs font-semibold text-[hsl(258_65%_40%)] shadow-sm transition-colors hover:bg-[hsl(258_75%_98%)]"
            >
              Open GitHub
            </a>
            <button
              type="button"
              onClick={() => void handleSponsor()}
              className="flex w-full items-center justify-center gap-2 rounded-full border border-white/50 px-4 py-1.5 text-xs font-semibold text-white transition-colors hover:bg-white/15"
            >
              <Heart size={14} weight="fill" />
              Sponsor DAWG
            </button>
          </div>
        </div>
      </aside>

      <main className="relative z-10 min-w-0 flex-1 overflow-y-auto bg-canvas">
        <div
          className={[
            "mx-auto flex w-full max-w-[1600px] flex-col px-5 pb-10 pt-7 sm:px-7 xl:px-9",
            activeTab === "artifacts" ? "h-full min-h-0" : "min-h-full",
          ].join(" ")}
        >
          <AnimatePresence mode="wait">
            <motion.div
              key={activeTab}
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ duration: 0.2, ease: "easeOut" }}
              className={
                activeTab === "artifacts" ? "flex min-h-0 flex-1 flex-col" : "flex flex-col"
              }
            >
              {activeTab === "dashboard" && (
                <Dashboard
                  onReplayArtifact={handleReplayArtifact}
                  onEditArtifact={handleEditArtifact}
                  onViewArtifacts={() => setActiveTab("artifacts")}
                  onStartCapture={() => setActiveTab("capture")}
                  onOpenReplay={() => setActiveTab("replay")}
                  onOpenEditor={() => setActiveTab("editor")}
                  onOpenPolicy={() => setActiveTab("policy")}
                  onOpenGuide={() => {
                    setActiveTab("dashboard");
                    setGuideStep(0);
                    setShowGuide(true);
                  }}
                />
              )}
              {activeTab === "capture" && <CaptureControls />}
              {activeTab === "artifacts" && (
                <ArtifactsPage
                  onReplayArtifact={handleReplayArtifact}
                  onEditArtifact={handleEditArtifact}
                />
              )}
              {activeTab === "replay" && (
                <ReplayViewer selectedArtifactIdentity={replayArtifactIdentity} />
              )}
              {activeTab === "editor" && (
                <EditorPage
                  selectedArtifactIdentity={editorArtifactIdentity}
                  onReplayArtifact={(identity) => {
                    setReplayArtifactIdentity(identity);
                    setActiveTab("replay");
                  }}
                />
              )}
              {activeTab === "policy" && <PolicyConfig />}
            </motion.div>
          </AnimatePresence>
        </div>
      </main>

      {showGuide && (
        <GuidedTour
          step={guideStep}
          activeTab={activeTab}
          onStepChange={setGuideStep}
          onNavigate={setActiveTab}
          onFinish={finishGuide}
        />
      )}

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
