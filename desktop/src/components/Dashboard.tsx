import { Archive, Cpu, Record, ShieldCheck } from "@phosphor-icons/react";
import type React from "react";
import { useEngine } from "../context/EngineContext";
import { ArtifactInspectorModal } from "./ArtifactInspectorModal";
import { ArtifactList } from "./ArtifactList";
import { EngineHelpBanner } from "./EngineHelpBanner";
import { LogStreamer } from "./LogStreamer";
import { Badge } from "./ui/Badge";
import { PageShell } from "./ui/PageShell";
import { StatCard } from "./ui/StatCard";

export const Dashboard: React.FC = () => {
  const { engineStatus, artifacts, inspectedArtifact, closeInspectModal, isCapturing } =
    useEngine();

  const today = new Date().toLocaleDateString("en-US", {
    weekday: "short",
    month: "long",
    day: "numeric",
  });

  return (
    <PageShell eyebrow={today} title="DAWG Workspace">
      {/* Engine warning banner (only shown when CLI missing) */}
      <EngineHelpBanner />

      {/* Capture live indicator */}
      {isCapturing && (
        <div className="flex items-center gap-3 p-4 bg-error-bg/50 border border-error-border rounded-xl shrink-0 shadow-sm backdrop-blur-md relative overflow-hidden">
          <div className="absolute top-0 left-0 w-1 h-full bg-error-text" />
          <Record size={18} weight="fill" className="text-error-text animate-pulse shrink-0" />
          <p className="text-sm font-semibold text-text-primary">
            Capture session is live — recording telemetry
          </p>
        </div>
      )}

      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        {/* Left Column: Artifacts */}
        <div className="xl:col-span-2 flex flex-col gap-4">
          <div className="flex items-center justify-between">
            <h3 className="text-base font-bold text-text-primary">Recent Artifacts</h3>
            <span className="text-xs font-semibold text-text-tertiary bg-canvas-subtle px-2 py-1 rounded-full border border-border">
              {artifacts.length} total
            </span>
          </div>
          <ArtifactList />

          <div className="mt-4">
            <h3 className="text-base font-bold text-text-primary mb-4">Engine Logs</h3>
            <LogStreamer />
          </div>
        </div>

        {/* Right Column: Stats & Policy */}
        <div className="flex flex-col gap-4">
          <h3 className="text-base font-bold text-text-primary">System Overview</h3>
          <StatCard
            label="Engine Status"
            icon={<Cpu size={18} weight="fill" />}
            value={
              <span className={engineStatus.installed ? "text-success-text" : "text-warning-text"}>
                {engineStatus.installed ? "Ready" : "Missing"}
              </span>
            }
            badge={
              engineStatus.installed ? (
                <Badge variant="success" dot size="sm">
                  {engineStatus.bundled ? "bundled" : "path"}
                </Badge>
              ) : (
                <Badge variant="warning" dot size="sm">
                  CLI not found
                </Badge>
              )
            }
          />

          <StatCard
            label="Total Captures"
            icon={<Archive size={18} weight="fill" />}
            value={artifacts.length}
            badge={
              artifacts.length > 0 ? (
                <Badge variant="brand" size="sm">
                  {artifacts.length} session{artifacts.length !== 1 ? "s" : ""}
                </Badge>
              ) : undefined
            }
          />

          <StatCard
            label="Sanitizer Policy"
            icon={<ShieldCheck size={18} weight="fill" />}
            value="default.rego"
            badge={
              <Badge variant="info" size="sm">
                OPA gate
              </Badge>
            }
          />
        </div>
      </div>

      <ArtifactInspectorModal artifact={inspectedArtifact} onClose={closeInspectModal} />
    </PageShell>
  );
};
