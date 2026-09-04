import { type ReactNode, createContext, useContext, useState } from "react";
import { type DoctorReport, type EngineStatusInfo, useEngineStatus } from "../hooks/useEngineStatus";
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
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);
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
      addLogLine(`Capture session active. Session ID: ${res.sessionId}`);
      if (res.sessionPath) {
        addLogLine(`Session path: ${res.sessionPath}`);
      }
    } catch (err) {
      addLogLine(`[ERROR] Start capture failed: ${String(err)}`);
      setIsCapturing(false);
    }
  };

  const stopCaptureSession = async () => {
    try {
      addLogLine("Stopping active capture session...");
      const sessionId = activeSessionId;
      const res = await engine.stopCapture();
      addLogLine(`Capture stop request sent: ${res.status} (control: ${res.controlFile})`);
      setIsCapturing(false);

      if (sessionId) {
        const artifactPath = `.dawg/artifacts/${sessionId}`;
        addLogLine(`Packaging complete. Artifact: ${artifactPath}`);
        setArtifacts((prev) => [
          {
            id: sessionId,
            path: artifactPath,
            targetUrl,
            createdAt: new Date().toISOString(),
            components: ["browser", "http", "cassettes", "actions"],
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
        doctorReport,
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
