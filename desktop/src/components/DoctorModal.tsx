import { ArrowsClockwise, CheckCircle, XCircle } from "@phosphor-icons/react";
/**
 * DoctorModal — DAWG Engine diagnostics dialog.
 * Shows engine path, resource root, and component status table.
 */
import { useEngine } from "../context/EngineContext";
import { Badge } from "./ui/Badge";
import { Button } from "./ui/Button";
import { Modal } from "./ui/Modal";

interface DoctorModalProps {
  open: boolean;
  onClose: () => void;
}

export function DoctorModal({ open, onClose }: DoctorModalProps) {
  // Touch file to resolve TS caching issue
  const { doctorReport, refetchEngineStatus, isCheckingEngine } = useEngine();

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Engine Doctor"
      subtitle="Runtime diagnostics and bundled resource verification"
      maxWidth="md"
      footerLeft={
        <Button
          variant="secondary"
          size="sm"
          loading={isCheckingEngine}
          iconLeft={<ArrowsClockwise size={13} />}
          onClick={refetchEngineStatus}
        >
          {isCheckingEngine ? "Re-probing…" : "Re-run Doctor"}
        </Button>
      }
    >
      {doctorReport ? (
        <div className="flex flex-col gap-4">
          {/* Status banner */}
          <div className="flex items-center gap-2">
            <Badge variant={doctorReport.status === "ready" ? "success" : "warning"} dot size="md">
              {doctorReport.status === "ready" ? "System Ready" : "Degraded"}
            </Badge>
            {doctorReport.isBundled && (
              <Badge variant="brand" size="sm">
                Bundled App Mode
              </Badge>
            )}
          </div>

          {/* Path info grid */}
          <div className="grid grid-cols-2 gap-3 p-3 bg-canvas-subtle rounded-md border border-border text-xs">
            <div>
              <span className="block text-[10px] text-text-tertiary uppercase tracking-wider mb-0.5">
                Engine Path
              </span>
              <span className="font-mono text-text-brand break-all">{doctorReport.enginePath}</span>
            </div>
            <div>
              <span className="block text-[10px] text-text-tertiary uppercase tracking-wider mb-0.5">
                Resource Root
              </span>
              <span className="font-mono text-text-brand break-all">
                {doctorReport.resourceDir}
              </span>
            </div>
          </div>

          {/* Component list */}
          <div className="flex flex-col gap-1">
            <span className="text-xs font-semibold text-text-secondary">Runtime Components</span>
            {doctorReport.components.map((c) => (
              <div
                key={c.name}
                className="flex items-center justify-between gap-3 p-3 rounded-md bg-canvas-subtle border border-border"
              >
                <div className="flex items-center gap-2.5 min-w-0">
                  {c.installed ? (
                    <CheckCircle size={16} weight="fill" className="text-success-dot shrink-0" />
                  ) : (
                    <XCircle size={16} weight="fill" className="text-error-dot shrink-0" />
                  )}
                  <div className="min-w-0">
                    <div className="flex items-center gap-1.5 text-sm font-medium text-text-primary">
                      <span>{c.name}</span>
                      {c.bundled && (
                        <Badge variant="info" size="sm">
                          bundled
                        </Badge>
                      )}
                      {c.version && (
                        <span className="font-mono text-xs text-text-tertiary">v{c.version}</span>
                      )}
                    </div>
                    {c.path && (
                      <div className="text-[10px] text-text-tertiary font-mono truncate max-w-xs mt-0.5">
                        {c.path}
                      </div>
                    )}
                    {c.error && <div className="text-xs text-error-text mt-0.5">{c.error}</div>}
                  </div>
                </div>
                <Badge variant={c.installed ? "success" : "error"} size="sm" dot>
                  {c.installed ? "Ready" : "Missing"}
                </Badge>
              </div>
            ))}
          </div>
        </div>
      ) : (
        <div className="py-12 text-center text-text-tertiary text-sm">
          {isCheckingEngine
            ? "Running diagnostics…"
            : "No diagnostic report available. Try re-running the doctor."}
        </div>
      )}
    </Modal>
  );
}
