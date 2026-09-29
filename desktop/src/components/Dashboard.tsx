import { Archive, ArrowRight, Cpu, Flag, Record, ShieldCheck } from "@phosphor-icons/react";
import type React from "react";
import { useMemo } from "react";

import { type ArtifactItem, useEngine } from "../context/EngineContext";
import { EngineHelpBanner } from "./EngineHelpBanner";
import { Badge } from "./ui/Badge";
import { Button } from "./ui/Button";
import { Card, CardHeader, CardTitle } from "./ui/Card";
import { EmptyState } from "./ui/EmptyState";
import { PageShell } from "./ui/PageShell";
import { StatCard } from "./ui/StatCard";

interface DashboardProps {
  onReplayArtifact: (artifact: ArtifactItem) => void;
  onEditArtifact: (artifact: ArtifactItem) => void;
  onViewArtifacts: () => void;
}

const addedTimestamp = (artifact: ArtifactItem) => {
  const value = new Date(artifact.addedAt || artifact.createdAt).getTime();
  return Number.isFinite(value) ? value : 0;
};

export const Dashboard: React.FC<DashboardProps> = ({
  onReplayArtifact,
  onEditArtifact,
  onViewArtifacts,
}) => {
  const { engineStatus, artifacts, doctorReport, isCapturing } = useEngine();
  const latestArtifact = useMemo(
    () => [...artifacts].sort((left, right) => addedTimestamp(right) - addedTimestamp(left))[0],
    [artifacts],
  );
  const summaries = useMemo(() => {
    const imported = artifacts.filter((artifact) => artifact.origin === "imported").length;
    const flagged = artifacts.filter((artifact) => artifact.flagged).length;
    const captured = artifacts.filter((artifact) => artifact.origin === "captured").length;
    const diagnostics = artifacts.reduce(
      (total, artifact) => total + (artifact.diagnosticsSummary?.total ?? 0),
      0,
    );
    return { imported, flagged, captured, diagnostics };
  }, [artifacts]);

  const compatibility = doctorReport?.compatibility;

  return (
    <PageShell
      eyebrow={new Date().toLocaleDateString("en-US", {
        weekday: "short",
        month: "long",
        day: "numeric",
      })}
      title="DAWG Workspace"
    >
      <EngineHelpBanner />
      {isCapturing && (
        <div className="relative flex items-center gap-3 overflow-hidden rounded-xl border border-error-border bg-error-bg/50 p-4 shadow-sm backdrop-blur-md">
          <div className="absolute top-0 left-0 h-full w-1 bg-error-text" />
          <Record size={18} weight="fill" className="shrink-0 animate-pulse text-error-text" />
          <p className="text-sm font-semibold text-text-primary">
            Capture session is live — recording telemetry
          </p>
        </div>
      )}

      <div className="grid grid-cols-1 gap-6 xl:grid-cols-3">
        <div className="flex flex-col gap-4 xl:col-span-2">
          <Card>
            <CardHeader>
              <CardTitle
                title="Latest artifact"
                subtitle="Most recently added local artifact instance"
              />
              <Button
                variant="secondary"
                size="sm"
                onClick={onViewArtifacts}
                iconRight={<ArrowRight size={13} />}
              >
                View all artifacts
              </Button>
            </CardHeader>
            {latestArtifact ? (
              <div className="flex flex-col gap-4">
                <div className="flex flex-wrap items-center gap-2">
                  <h3 className="min-w-0 truncate text-base font-bold text-text-primary">
                    {latestArtifact.title || latestArtifact.id}
                  </h3>
                  {latestArtifact.origin === "imported" && (
                    <Badge variant="info" size="sm">
                      Imported
                    </Badge>
                  )}
                  {latestArtifact.flagged && (
                    <Badge variant="warning" size="sm">
                      Flagged{latestArtifact.flagCount ? ` · ${latestArtifact.flagCount}` : ""}
                    </Badge>
                  )}
                  {latestArtifact.revision && (
                    <Badge variant="brand" size="sm">
                      Revision {latestArtifact.revision}
                    </Badge>
                  )}
                </div>
                <div className="grid grid-cols-1 gap-3 rounded-md border border-border bg-canvas-subtle p-3 text-xs sm:grid-cols-2">
                  <div>
                    <span className="block text-text-tertiary">Target URL</span>
                    <span className="font-medium text-text-primary wrap-anywhere">
                      {latestArtifact.targetUrl || "Not recorded"}
                    </span>
                  </div>
                  <div>
                    <span className="block text-text-tertiary">Added</span>
                    <span className="font-medium text-text-primary">
                      {new Date(
                        latestArtifact.addedAt || latestArtifact.createdAt,
                      ).toLocaleString()}
                    </span>
                  </div>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Button size="sm" onClick={() => onReplayArtifact(latestArtifact)}>
                    Replay
                  </Button>
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => onEditArtifact(latestArtifact)}
                  >
                    Edit review
                  </Button>
                </div>
              </div>
            ) : (
              <EmptyState
                icon={<Archive size={28} weight="light" />}
                title="No artifacts yet"
                description="Capture a session or import a .dawg archive from the Artifacts page."
              />
            )}
          </Card>

          <Card>
            <CardHeader>
              <CardTitle
                title="Recent activity"
                subtitle="Catalog information available without opening every artifact"
              />
            </CardHeader>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
              <Summary
                label="Recent captures"
                value={summaries.captured}
                detail="Local acquisition instances"
              />
              <Summary
                label="Review flags"
                value={summaries.flagged}
                detail="Flagged artifact instances"
              />
              <Summary
                label="Diagnostics"
                value={summaries.diagnostics}
                detail="Recorded summary total"
              />
            </div>
          </Card>
        </div>

        <div className="flex flex-col gap-4">
          <h3 className="text-base font-bold text-text-primary">System overview</h3>
          <StatCard
            label="Engine status"
            icon={<Cpu size={18} weight="fill" />}
            value={
              <span className={engineStatus.installed ? "text-success-text" : "text-warning-text"}>
                {engineStatus.installed ? "Ready" : "Missing"}
              </span>
            }
            badge={
              <Badge variant={engineStatus.installed ? "success" : "warning"} dot size="sm">
                {engineStatus.installed
                  ? engineStatus.bundled
                    ? "bundled"
                    : "path"
                  : "CLI not found"}
              </Badge>
            }
          />
          <StatCard
            label="Local artifacts"
            icon={<Archive size={18} weight="fill" />}
            value={artifacts.length}
            badge={
              <Badge variant="brand" size="sm">
                {summaries.imported} imported
              </Badge>
            }
          />
          <StatCard
            label="Doctor compatibility"
            icon={<ShieldCheck size={18} weight="fill" />}
            value={
              <span
                className={
                  compatibility?.status === "compatible" ? "text-success-text" : "text-warning-text"
                }
              >
                {compatibility?.status === "compatible"
                  ? "Compatible"
                  : compatibility?.status === "mismatch"
                    ? "Mismatch"
                    : "Unknown"}
              </span>
            }
            badge={
              <Badge
                variant={compatibility?.status === "compatible" ? "success" : "warning"}
                size="sm"
              >
                {compatibility?.expectedApplicationVersion || "Run Doctor"}
              </Badge>
            }
          />
          <StatCard
            label="Flagged artifacts"
            icon={<Flag size={18} weight="fill" />}
            value={summaries.flagged}
            badge={
              <Badge variant="info" size="sm">
                Portable review
              </Badge>
            }
          />
        </div>
      </div>
    </PageShell>
  );
};

function Summary({ label, value, detail }: { label: string; value: number; detail: string }) {
  return (
    <div className="rounded-md border border-border bg-canvas-subtle p-3">
      <span className="block text-xs text-text-tertiary">{label}</span>
      <strong className="mt-1 block text-xl text-text-primary">{value}</strong>
      <span className="text-[11px] text-text-tertiary">{detail}</span>
    </div>
  );
}
