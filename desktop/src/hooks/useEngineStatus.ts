import { useCallback, useEffect, useState } from "react";
import { type DoctorReport, type EngineStatusInfo, engine } from "../lib/engine";

export type { EngineStatusInfo, DoctorReport };

export function useEngineStatus() {
  const [status, setStatus] = useState<EngineStatusInfo>({
    installed: false,
    bundled: false,
    error: "Initializing engine status probe...",
  });
  const [doctorReport, setDoctorReport] = useState<DoctorReport | null>(null);
  const [isChecking, setIsChecking] = useState(true);

  const checkStatus = useCallback(async () => {
    setIsChecking(true);
    try {
      const info = await engine.checkInstalled();
      setStatus(info);

      // Attempt to load full doctor diagnostics
      try {
        const report = await engine.getDoctorReport();
        setDoctorReport(report);
      } catch (docErr) {
        console.warn("Doctor report probe note:", docErr);
      }
    } catch (err) {
      setStatus({
        installed: false,
        bundled: false,
        error: `Tauri IPC check failed: ${String(err)}`,
      });
    } finally {
      setIsChecking(false);
    }
  }, []);

  useEffect(() => {
    checkStatus();
  }, [checkStatus]);

  return { status, doctorReport, isChecking, refetch: checkStatus };
}
