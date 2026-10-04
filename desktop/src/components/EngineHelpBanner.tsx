import { ArrowsClockwise, WarningCircle } from "@phosphor-icons/react";
import type React from "react";
import { useEngine } from "../context/EngineContext";
import { Button } from "./ui/Button";
import { Card } from "./ui/Card";

export const EngineHelpBanner: React.FC = () => {
  const { engineStatus, refetchEngineStatus, isCheckingEngine } = useEngine();

  if (engineStatus.installed) {
    return null;
  }

  return (
    <Card className="!border-warning-border !bg-warning-bg">
      <div className="flex flex-col gap-3">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-center gap-2">
            <WarningCircle size={21} weight="fill" className="text-warning-text" />
            <h3 className="text-sm font-semibold text-warning-text">DAWG engine not found</h3>
          </div>
          <Button
            variant="secondary"
            size="sm"
            onClick={refetchEngineStatus}
            loading={isCheckingEngine}
            iconLeft={!isCheckingEngine && <ArrowsClockwise size={14} />}
            className="shrink-0 !border-warning-border !bg-surface !text-warning-text hover:!bg-warning-border hover:!text-white"
          >
            {isCheckingEngine ? "Checking..." : "Check again"}
          </Button>
        </div>

        <p className="text-xs text-warning-text">Install the DAWG CLI to capture and replay.</p>

        <details className="text-xs text-warning-text">
          <summary className="w-fit cursor-pointer font-semibold">Show setup steps</summary>
          <div className="mt-3 space-y-1 overflow-x-auto rounded-xl border border-warning-border/50 bg-surface p-3 font-mono">
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
        </details>
      </div>
    </Card>
  );
};
