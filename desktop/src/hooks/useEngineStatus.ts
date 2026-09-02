import { invoke } from "@tauri-apps/api/core";
import { useCallback, useEffect, useState } from "react";

export interface EngineStatusInfo {
  installed: boolean;
  version?: string;
  path?: string;
  error?: string;
}

export function useEngineStatus() {
  const [status, setStatus] = useState<EngineStatusInfo>({
    installed: false,
    error: "Initializing engine status probe...",
  });
  const [isChecking, setIsChecking] = useState(true);

  const checkStatus = useCallback(async () => {
    setIsChecking(true);
    try {
      const info = await invoke<EngineStatusInfo>("check_engine_installed");
      setStatus(info);
    } catch (err) {
      setStatus({
        installed: false,
        error: `Tauri IPC check failed: ${String(err)}`,
      });
    } finally {
      setIsChecking(false);
    }
  }, []);

  useEffect(() => {
    checkStatus();
  }, [checkStatus]);

  return { status, isChecking, refetch: checkStatus };
}
