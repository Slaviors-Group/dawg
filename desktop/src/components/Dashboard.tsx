import {
  Archive,
  ArrowRight,
  ArrowUpRight,
  Cpu,
  Flag,
  Record,
  ShieldCheck,
  StarFour,
} from "@phosphor-icons/react";
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
  onStartCapture: () => void;
}

const addedTimestamp = (artifact: ArtifactItem) => {
  const value = new Date(artifact.addedAt || artifact.createdAt).getTime();
  return Number.isFinite(value) ? value : 0;
};

export const Dashboard: React.FC<DashboardProps> = ({
  onReplayArtifact,
  onEditArtifact,
  onViewArtifacts,
  onStartCapture,
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
  const otherArtifacts = Math.max(0, artifacts.length - summaries.captured - summaries.imported);
  const highestSourceCount = Math.max(1, summaries.captured, summaries.imported, otherArtifacts);

  return (
    <PageShell
      eyebrow={new Date().toLocaleDateString("en-US", {
        weekday: "long",
        month: "long",
        day: "numeric",
      })}
      title="Dashboard"
      subtitle="Capture, inspect, and replay browser issues."
      actions={
        <Badge variant={engineStatus.installed ? "success" : "warning"} dot>
          {engineStatus.installed ? "Engine ready" : "Engine unavailable"}
        </Badge>
      }
    >
      <EngineHelpBanner />
      {isCapturing && (
        <div className="flex items-center gap-3 rounded-2xl border border-error-border bg-error-bg px-4 py-3 text-sm font-medium text-error-text">
          <Record size={17} weight="fill" className="shrink-0 animate-pulse" />
          Capture is live. Stop it from the Capture page when finished.
        </div>
      )}

      <div className="grid grid-cols-1 gap-5 xl:grid-cols-[minmax(0,1.8fr)_minmax(280px,1fr)]">
        <section className="relative flex min-h-[270px] flex-col justify-between overflow-hidden rounded-[26px] bg-brand-500 p-6 text-white sm:p-8">
          <div
            className="pointer-events-none absolute -right-10 -top-28 size-80 rounded-full border border-white/15"
            aria-hidden="true"
          />
          <div
            className="pointer-events-none absolute -right-28 -top-44 size-[28rem] rounded-full border border-white/10"
            aria-hidden="true"
          />
          <StarFour
            size={70}
            weight="thin"
            className="pointer-events-none absolute right-[19%] top-[16%] text-white/45"
            aria-hidden="true"
          />
          <StarFour
            size={35}
            weight="thin"
            className="pointer-events-none absolute right-[8%] bottom-[15%] text-white/30"
            aria-hidden="true"
          />
          <div className="relative z-10">
            <span className="text-[11px] font-semibold uppercase tracking-[0.2em] text-white/80">
              DAWG workspace
            </span>
            <h2 className="mt-4 max-w-xl text-[2rem] font-semibold leading-[1.1] tracking-tight sm:text-[2.5rem]">
              Find the issue. Keep the evidence.
            </h2>
            <p className="mt-3 max-w-md text-sm leading-relaxed text-white/85">
              Turn browser sessions into artifacts your team can inspect and replay.
            </p>
          </div>
          <div className="relative z-10 mt-6 flex flex-wrap items-center gap-2">
            <Button
              onClick={onStartCapture}
              iconRight={<ArrowUpRight size={16} weight="bold" />}
              className="!bg-[#1c1b2b] !text-white hover:!bg-[#33314a]"
            >
              Start capture
            </Button>
            <Button
              variant="ghost"
              onClick={onViewArtifacts}
              className="!text-white hover:!bg-white/15"
            >
              Browse artifacts
            </Button>
          </div>
        </section>

        <Card className="flex flex-col">
          <CardHeader>
            <CardTitle title="Workspace status" />
            <span className="grid size-9 place-items-center rounded-full bg-brand-100 text-brand-600">
              <Cpu size={18} weight="duotone" />
            </span>
          </CardHeader>
          <div className="flex items-center justify-between rounded-xl bg-canvas-subtle px-4 py-3">
            <div>
              <p className="text-xs text-text-tertiary">Engine</p>
              <p className="mt-0.5 text-sm font-semibold text-text-primary">
                {engineStatus.installed ? "Ready to capture" : "Setup needed"}
              </p>
            </div>
            <Badge variant={engineStatus.installed ? "success" : "warning"} dot size="sm">
              {engineStatus.installed ? "Ready" : "Missing"}
            </Badge>
          </div>
          <div className="mt-5">
            <div className="mb-3 flex items-center justify-between">
              <span className="text-xs font-semibold text-text-secondary">Artifact sources</span>
              <span className="text-xs text-text-tertiary">{artifacts.length} total</span>
            </div>
            <div className="space-y-3">
              <SourceRow label="Captured" count={summaries.captured} max={highestSourceCount} />
              <SourceRow label="Imported" count={summaries.imported} max={highestSourceCount} />
              <SourceRow label="Other" count={otherArtifacts} max={highestSourceCount} />
            </div>
          </div>
          <div className="mt-auto flex items-center justify-between border-t border-border/70 pt-4 text-xs">
            <span className="text-text-tertiary">Doctor check</span>
            <span className="font-semibold text-text-primary">
              {compatibility?.status === "compatible"
                ? "Compatible"
                : compatibility?.status === "mismatch"
                  ? "Mismatch"
                  : "Not checked"}
            </span>
          </div>
        </Card>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard
          label="Local artifacts"
          icon={<Archive size={20} weight="duotone" />}
          value={artifacts.length}
          badge={
            <Badge variant="brand" size="sm">
              {summaries.imported} imported
            </Badge>
          }
        />
        <StatCard
          label="Review flags"
          icon={<Flag size={20} weight="duotone" />}
          value={summaries.flagged}
          badge={<span className="text-xs text-text-tertiary">Needs review</span>}
        />
        <StatCard
          label="Diagnostics"
          icon={<ShieldCheck size={20} weight="duotone" />}
          value={summaries.diagnostics}
          badge={<span className="text-xs text-text-tertiary">Recorded evidence</span>}
        />
      </div>

      <div className="grid grid-cols-1 gap-5 xl:grid-cols-[minmax(0,1.8fr)_minmax(280px,1fr)]">
        <Card>
          <CardHeader>
            <CardTitle title="Latest artifact" />
            <Button
              variant="ghost"
              size="sm"
              onClick={onViewArtifacts}
              iconRight={<ArrowRight size={14} />}
            >
              View all
            </Button>
          </CardHeader>
          {latestArtifact ? (
            <div className="flex flex-col gap-4">
              <div className="flex flex-wrap items-center gap-2">
                <h3 className="min-w-0 truncate text-lg font-semibold text-text-primary">
                  {latestArtifact.title || latestArtifact.id}
                </h3>
                {latestArtifact.origin === "imported" && (
                  <Badge variant="info" size="sm">
                    Imported
                  </Badge>
                )}
                {latestArtifact.flagged && (
                  <Badge variant="warning" size="sm">
                    Flagged
                  </Badge>
                )}
                {latestArtifact.revision && (
                  <Badge variant="brand" size="sm">
                    Rev {latestArtifact.revision}
                  </Badge>
                )}
              </div>
              <div className="grid grid-cols-1 gap-3 rounded-xl bg-canvas-subtle p-4 text-sm sm:grid-cols-2">
                <div className="min-w-0">
                  <span className="block text-xs text-text-tertiary">Target</span>
                  <span
                    className="mt-1 block truncate font-medium text-text-primary"
                    title={latestArtifact.targetUrl}
                  >
                    {latestArtifact.targetUrl || "Not recorded"}
                  </span>
                </div>
                <div>
                  <span className="block text-xs text-text-tertiary">Added</span>
                  <span className="mt-1 block font-medium text-text-primary">
                    {new Date(latestArtifact.addedAt || latestArtifact.createdAt).toLocaleString()}
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
              icon={<Archive size={28} weight="regular" />}
              title="No artifacts yet"
              description="Capture a session or import a DAWG archive."
            />
          )}
        </Card>

        <Card>
          <CardHeader>
            <CardTitle title="Quick actions" />
          </CardHeader>
          <div className="space-y-2">
            <button
              type="button"
              onClick={onStartCapture}
              className="flex w-full items-center gap-3 rounded-xl border border-border/70 bg-canvas-subtle px-4 py-3 text-left transition-colors hover:border-brand-300 hover:bg-brand-50"
            >
              <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-brand-100 text-brand-600">
                <Record size={19} />
              </span>
              <span className="min-w-0 flex-1">
                <span className="block text-sm font-semibold text-text-primary">
                  Capture a session
                </span>
                <span className="block text-xs text-text-tertiary">Record a browser issue</span>
              </span>
              <ArrowUpRight size={16} className="shrink-0 text-text-tertiary" />
            </button>
            <button
              type="button"
              onClick={onViewArtifacts}
              className="flex w-full items-center gap-3 rounded-xl border border-border/70 bg-canvas-subtle px-4 py-3 text-left transition-colors hover:border-brand-300 hover:bg-brand-50"
            >
              <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-brand-100 text-brand-600">
                <Archive size={19} />
              </span>
              <span className="min-w-0 flex-1">
                <span className="block text-sm font-semibold text-text-primary">Open catalog</span>
                <span className="block text-xs text-text-tertiary">
                  Inspect and export artifacts
                </span>
              </span>
              <ArrowUpRight size={16} className="shrink-0 text-text-tertiary" />
            </button>
          </div>
        </Card>
      </div>
    </PageShell>
  );
};

function SourceRow({ label, count, max }: { label: string; count: number; max: number }) {
  return (
    <div className="grid grid-cols-[72px_minmax(0,1fr)_24px] items-center gap-3 text-xs">
      <span className="text-text-secondary">{label}</span>
      <span className="h-2 overflow-hidden rounded-full bg-brand-100 dark:bg-surface-hover">
        <span
          className="block h-full rounded-full bg-brand-500"
          style={{ width: `${(count / max) * 100}%` }}
        />
      </span>
      <span className="text-right font-semibold text-text-primary">{count}</span>
    </div>
  );
}
