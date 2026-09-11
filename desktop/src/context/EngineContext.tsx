import { type ReactNode, createContext, useContext, useState } from "react";
import {
  type DoctorReport,
  type EngineStatusInfo,
  useEngineStatus,
} from "../hooks/useEngineStatus";
import { engine } from "../lib/engine";

export interface ArtifactItem {
  id: string;
  path: string;
  targetUrl: string;
  createdAt: string;
  components: string[];
}

interface EngineContextType {
  engineStatus: EngineStatusInfo;
  doctorReport: DoctorReport | null;
  isCheckingEngine: boolean;
  refetchEngineStatus: () => Promise<void>;
  isCapturing: boolean;
  isStopping: boolean;
  activeSessionId: string | null;
  targetUrl: string;
  setTargetUrl: (url: string) => void;
  logs: string[];
  artifacts: ArtifactItem[];
  inspectedArtifact: ArtifactItem | null;
  openInspectModal: (artifact: ArtifactItem) => void;
  closeInspectModal: () => void;
  startCaptureSession: (url: string) => Promise<void>;
  stopCaptureSession: () => Promise<void>;
  addLogLine: (line: string) => void;
  clearLogs: () => void;
}

const EngineContext = createContext<EngineContextType | undefined>(undefined);

export function EngineProvider({ children }: { children: ReactNode }) {
  const {
    status: engineStatus,
    doctorReport,
    isChecking: isCheckingEngine,
    refetch: refetchEngineStatus,
  } = useEngineStatus();
  const [isCapturing, setIsCapturing] = useState(false);
  const [isStopping, setIsStopping] = useState(false);
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);
  const [activeControlFile, setActiveControlFile] = useState<string | null>(null);
  const [targetUrl, setTargetUrl] = useState("http://localhost:3000");
  const [inspectedArtifact, setInspectedArtifact] = useState<ArtifactItem | null>(null);
  const [logs, setLogs] = useState<string[]>([
    "[SYSTEM] DAWG Desktop Shell initialized.",
    "[SYSTEM] Engine bridge initialized.",
  ]);
  const [artifacts, setArtifacts] = useState<ArtifactItem[]>([]);

  const addLogLine = (line: string) => {
    const timestamp = new Date().toLocaleTimeString();
    setLogs((prev) => [...prev, `[${timestamp}] ${line}`]);
  };

  const clearLogs = () => {
    setLogs([]);
  };

  const openInspectModal = (artifact: ArtifactItem) => {
    setInspectedArtifact(artifact);
  };

  const closeInspectModal = () => {
    setInspectedArtifact(null);
  };

  const startCaptureSession = async (url: string) => {
    try {
      addLogLine(`Starting capture session for target URL: ${url}`);
      setIsCapturing(true);
      const res = await engine.startCapture({ url });
      setActiveSessionId(res.sessionId);
      setActiveControlFile(res.controlFile);
      addLogLine(`Capture session active. Session ID: ${res.sessionId}`);
      if (res.sessionPath) {
        addLogLine(`Session path: ${res.sessionPath}`);
      }
    } catch (err) {
      addLogLine(`[ERROR] Start capture failed: ${String(err)}`);
      setIsCapturing(false);
      setActiveControlFile(null);
    }
  };

  const stopCaptureSession = async () => {
    // `capture stop` now blocks in the engine until sanitize + packaging
    // finish, so it can hand back the real artifact path instead of a
    // guessed one. That means this can legitimately take several seconds.
    setIsStopping(true);
    try {
      addLogLine("Stopping active capture session and packaging artifact...");
      const res = await engine.stopCapture({
        controlFile: activeControlFile ?? undefined,
      });
      setIsCapturing(false);

      const sessionId = res.sessionId ?? activeSessionId;
      if (sessionId && res.artifactPath) {
        addLogLine(`Packaging complete. Artifact: ${res.artifactPath}`);
        setArtifacts((prev) => [
          {
            id: sessionId,
            path: res.artifactPath as string,
            targetUrl,
            createdAt: new Date().toISOString(),
            components: ["browser", "http", "actions"],
          },
          ...prev,
        ]);
      } else {
        addLogLine("[WARN] Capture stopped but no artifact path was returned.");
      }
      setActiveSessionId(null);
      setActiveControlFile(null);
    } catch (err) {
      addLogLine(`[ERROR] Stop capture failed: ${String(err)}`);
      // Reset state so the user isn't permanently stuck — the daemon is already
      // gone if we get ECONNREFUSED, so treat this as an abnormal session end.
      setIsCapturing(false);
      setActiveSessionId(null);
      setActiveControlFile(null);
    } finally {
      setIsStopping(false);
    }
  };

  return (
    <EngineContext.Provider
      value={{
        engineStatus,
        doctorReport,
        isCheckingEngine,
        refetchEngineStatus,
        isCapturing,
        isStopping,
        activeSessionId,
        targetUrl,
        setTargetUrl,
        logs,
        artifacts,
        inspectedArtifact,
        openInspectModal,
        closeInspectModal,
        startCaptureSession,
        stopCaptureSession,
        addLogLine,
        clearLogs,
      }}
    >
      {children}
    </EngineContext.Provider>
  );
}

export function useEngine() {
  const context = useContext(EngineContext);
  if (!context) {
    throw new Error("useEngine must be used within an EngineProvider");
  }
  return context;
}
