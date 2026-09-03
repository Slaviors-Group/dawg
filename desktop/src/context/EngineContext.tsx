import { type ReactNode, createContext, useContext, useState } from "react";
import { type EngineStatusInfo, useEngineStatus } from "../hooks/useEngineStatus";
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
  isCheckingEngine: boolean;
  refetchEngineStatus: () => Promise<void>;
  isCapturing: boolean;
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
    isChecking: isCheckingEngine,
    refetch: refetchEngineStatus,
  } = useEngineStatus();
  const [isCapturing, setIsCapturing] = useState(false);
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);
  const [targetUrl, setTargetUrl] = useState("http://localhost:3000");
  const [inspectedArtifact, setInspectedArtifact] = useState<ArtifactItem | null>(null);
  const [logs, setLogs] = useState<string[]>([
    "[SYSTEM] DAWG Desktop Shell initialized.",
    "[SYSTEM] Engine bridge initialized in passive mode.",
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
      addLogLine(`Capture session active. Session ID: ${res.sessionId}`);
    } catch (err) {
      addLogLine(`[ERROR] Start capture failed: ${String(err)}`);
      setIsCapturing(false);
    }
  };

  const stopCaptureSession = async () => {
    try {
      addLogLine("Stopping active capture session...");
      const res = await engine.stopCapture();
      addLogLine(`Capture stopped. Artifact created at: ${res.artifactPath}`);
      setIsCapturing(false);

      if (activeSessionId) {
        setArtifacts((prev) => [
          {
            id: activeSessionId,
            path: res.artifactPath,
            targetUrl,
            createdAt: new Date().toISOString(),
            components: ["browser", "http", "dbDiff", "logs"],
          },
          ...prev,
        ]);
      }
      setActiveSessionId(null);
    } catch (err) {
      addLogLine(`[ERROR] Stop capture failed: ${String(err)}`);
    }
  };

  return (
    <EngineContext.Provider
      value={{
        engineStatus,
        isCheckingEngine,
        refetchEngineStatus,
        isCapturing,
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
