import type React from "react";
import { useEngine } from "../context/EngineContext";
import { Card } from "./ui/Card";
import { Button } from "./ui/Button";
import { WarningCircle, ArrowsClockwise } from "@phosphor-icons/react";

export const EngineHelpBanner: React.FC = () => {
  const { engineStatus, refetchEngineStatus, isCheckingEngine } = useEngine();

  if (engineStatus.installed) {
    return null;
  }

  return (
    <Card className="!bg-warning-bg !border-warning-border">
      <div className="flex flex-col gap-4">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-center gap-2">
            <WarningCircle
              size={24}
              weight="fill"
              className="text-warning-text"
            />
            <h3 className="text-sm font-semibold text-warning-text">
              DAWG Engine CLI Binary Not Found on PATH
            </h3>
          </div>
          <Button
            variant="secondary"
            size="sm"
            onClick={refetchEngineStatus}
            loading={isCheckingEngine}
            iconLeft={!isCheckingEngine && <ArrowsClockwise size={14} />}
            className="!bg-surface !text-warning-text !border-warning-border hover:!bg-warning-border hover:!text-white shrink-0"
          >
            {isCheckingEngine ? "Probing..." : "Re-probe PATH"}
          </Button>
        </div>

        <p className="text-xs text-warning-text/80 leading-relaxed">
          The Desktop Shell uses Tauri IPC to spawn `dawg` CLI subprocesses. To build and install the
          engine binary locally:
        </p>

        <div className="bg-warning-border/30 border border-warning-border/50 p-3 rounded-md font-mono text-xs text-warning-text overflow-x-auto space-y-1">
          <div>
            <span className="opacity-60"># 1. Navigate to engine directory</span>
          </div>
          <div>
            <span className="font-semibold">cd</span> dawg/engine
          </div>
          <div className="mt-2">
            <span className="opacity-60"># 2. Build the CLI binary</span>
          </div>
          <div>
            <span className="font-semibold">go build</span> -o dawg ./cmd/dawg
          </div>
          <div className="mt-2">
            <span className="opacity-60"># 3. Add to system PATH or execute dawg init</span>
          </div>
        </div>
      </div>
    </Card>
  );
};
