import {
  Cube,
  Export,
  Flask,
  Info,
  MagnifyingGlass,
  PlayCircle,
  StopCircle,
  UploadSimple,
} from "@phosphor-icons/react";
import { listen } from "@tauri-apps/api/event";
import type React from "react";
import { useEffect, useMemo, useRef, useState } from "react";
import { useEngine } from "../context/EngineContext";
import {
  chooseArtifactArchive,
  chooseArtifactExportPath,
  chooseDiagnosticHARPath,
  chooseDiagnosticRemovalParentDirectory,
  defaultArtifactExportName,
} from "../lib/artifactDialogs";
import {
  type DiagnosticCategory,
  type DiagnosticEvidence,
  type InspectResult,
  type ReplayControlType,
  type ReplayEvent,
  type ReplaySpeed,
  engine,
} from "../lib/engine";
import { DiagnosticRemovalModal } from "./DiagnosticRemovalModal";
import { DiagnosticReviewModal } from "./DiagnosticReviewModal";
import { LogStreamer } from "./LogStreamer";
import { ReplayDiagnosticInspector } from "./ReplayDiagnosticInspector";
import { Badge } from "./ui/Badge";
import { Button } from "./ui/Button";
import { Card, CardHeader, CardTitle } from "./ui/Card";
import { EmptyState } from "./ui/EmptyState";
import { Input } from "./ui/Input";
import { PageShell } from "./ui/PageShell";
import { Select } from "./ui/Select";

// UI-only OS detection; the engine selects the actual replay mode.
const isWindows = typeof navigator !== "undefined" && navigator.userAgent.includes("Windows");
const REPLAY_SPEEDS: ReplaySpeed[] = [0.5, 1, 1.5, 2, 4];

const formatReplayTime = (milliseconds: number) => {
  const totalSeconds = Math.floor(Math.max(0, milliseconds) / 1000);
  const seconds = totalSeconds % 60;
  const totalMinutes = Math.floor(totalSeconds / 60);
  const minutes = totalMinutes % 60;
  const hours = Math.floor(totalMinutes / 60);
  const base = `${minutes}:${String(seconds).padStart(2, "0")}`;
  return hours > 0 ? `${hours}:${base.padStart(5, "0")}` : base;
};

interface ReplayViewerProps {
  selectedArtifactPath?: string;
}

export const ReplayViewer: React.FC<ReplayViewerProps> = ({ selectedArtifactPath }) => {
  const { artifacts, addLogLine, exportArtifact, importArtifact } = useEngine();
  const [selectedArtifact, setSelectedArtifact] = useState("");
  const [artifactSearch, setArtifactSearch] = useState("");
  const [originFilter, setOriginFilter] = useState("all");
  const [isReplaying, setIsReplaying] = useState(false);
  const [isCancelling, setIsCancelling] = useState(false);
  const [isImporting, setIsImporting] = useState(false);
  const [isExporting, setIsExporting] = useState(false);
  const [manifest, setManifest] = useState<InspectResult | null>(null);
  const [isLoadingManifest, setIsLoadingManifest] = useState(false);
  const [manifestError, setManifestError] = useState<string | null>(null);
  const [diagnosticEvidence, setDiagnosticEvidence] = useState<DiagnosticEvidence | null>(null);
  const [reviewAction, setReviewAction] = useState<"artifact" | "har" | "curl" | null>(null);
  const [reviewRequestId, setReviewRequestId] = useState<string | null>(null);
  const [isDiagnosticRemovalOpen, setIsDiagnosticRemovalOpen] = useState(false);
  const [replayReady, setReplayReady] = useState(false);
  const [replayPlaying, setReplayPlaying] = useState(false);
  const [replayCurrentTime, setReplayCurrentTime] = useState(0);
  const [replayDuration, setReplayDuration] = useState(0);
  const [replaySpeed, setReplaySpeed] = useState<ReplaySpeed>(1);
  const [seekPreview, setSeekPreview] = useState<number | null>(null);
  const replaySequence = useRef(0);
  const replayCommandSequence = useRef(0);

  useEffect(() => {
    let disposed = false;
    let unlisten: (() => void) | undefined;
    void listen<ReplayEvent>("dawg://replay-event", ({ payload }) => {
      if (payload.protocol !== "dawg.replay.v1" || payload.sequence <= replaySequence.current) return;
      replaySequence.current = payload.sequence;
      if (payload.type === "error") {
        addLogLine(`[ERROR] Replay control failed: ${payload.message ?? "unknown error"}`);
        return;
      }
      if (payload.type === "closed") {
        setIsReplaying(false);
        setReplayReady(false);
        setReplayPlaying(false);
        return;
      }
      if (payload.type === "ready" || payload.type === "state" || payload.type === "finished") {
        setIsReplaying(true);
        setReplayReady(true);
        setReplayPlaying(payload.type === "finished" ? false : Boolean(payload.playing));
        if (Number.isFinite(payload.currentTimeMs)) setReplayCurrentTime(payload.currentTimeMs ?? 0);
        if (Number.isFinite(payload.durationMs)) setReplayDuration(payload.durationMs ?? 0);
        if (payload.speed && REPLAY_SPEEDS.includes(payload.speed)) setReplaySpeed(payload.speed);
      }
    }).then((stopListening) => {
      if (disposed) {
        stopListening();
        return;
      }
      unlisten = stopListening;
      replayCommandSequence.current += 1;
      void engine
        .controlReplay({ id: `desktop-${replayCommandSequence.current}`, type: "getState" })
        .catch(() => {
          // No replay is expected during a normal initial mount.
        });
    });
    return () => {
      disposed = true;
      unlisten?.();
    };
  }, [addLogLine]);

  useEffect(() => {
    if (
      selectedArtifactPath &&
      artifacts.some((artifact) => artifact.path === selectedArtifactPath)
    ) {
      setSelectedArtifact(selectedArtifactPath);
    }
  }, [artifacts, selectedArtifactPath]);

  const selectedArtifactItem = useMemo(
    () => artifacts.find((artifact) => artifact.path === selectedArtifact),
    [artifacts, selectedArtifact],
  );

  useEffect(() => {
    let disposed = false;
    if (!selectedArtifact) {
      setManifest(null);
      setDiagnosticEvidence(null);
      setManifestError(null);
      setIsLoadingManifest(false);
      return () => {
        disposed = true;
      };
    }

    setManifest(null);
    setDiagnosticEvidence(null);
    setManifestError(null);
    setIsLoadingManifest(true);
    void Promise.all([
      engine.inspectArtifact({ path: selectedArtifact }),
      engine.inspectDiagnostics(selectedArtifact),
    ])
      .then(([result, evidence]) => {
        if (disposed) return;
        setManifest(result);
        setDiagnosticEvidence(evidence);
      })
      .catch((inspectError) => {
        if (!disposed) setManifestError(String(inspectError));
      })
      .finally(() => {
        if (!disposed) setIsLoadingManifest(false);
      });

    return () => {
      disposed = true;
    };
  }, [selectedArtifact]);

  const sandboxStatus = useMemo(() => {
    if (isWindows) {
      return {
        title: "Native Compatibility Mode (Windows)",
        description:
          "Bare Windows hosts skip rootless Docker isolation and DB fixture restore. Browser replay still runs natively. Use WSL2/Linux with rootless Docker for full sandbox isolation and DB restore.",
      };
    }
    return {
      title: "Rootless Docker Sandbox",
      description:
        "Replay isolates the captured environment in a rootless Docker Compose project. Run a replay to see live sandbox status in the execution logs.",
    };
  }, []);

  const visibleArtifacts = useMemo(() => {
    const query = artifactSearch.trim().toLowerCase();
    return artifacts.filter((artifact) => {
      if (originFilter !== "all" && artifact.origin !== originFilter) return false;
      if (!query) return true;
      return [artifact.title, artifact.id, artifact.targetUrl, artifact.path]
        .join(" ")
        .toLowerCase()
        .includes(query);
    });
  }, [artifactSearch, artifacts, originFilter]);

  const handleImportArtifact = async () => {
    if (isImporting) return;
    const archive = await chooseArtifactArchive();
    if (!archive) return;

    setIsImporting(true);
    try {
      await importArtifact(archive);
    } catch {
      // EngineContext records the actionable error in the shared log stream.
    } finally {
      setIsImporting(false);
    }
  };

  const exportArtifactAfterReview = async () => {
    if (!selectedArtifactItem || isExporting) return;
    const output = await chooseArtifactExportPath(
      defaultArtifactExportName(
        selectedArtifactItem.title || selectedArtifactItem.id,
        selectedArtifactItem.createdAt,
      ),
    );
    if (!output) return;

    setIsExporting(true);
    try {
      await exportArtifact(selectedArtifactItem, output);
    } catch {
      // EngineContext records the actionable error in the shared log stream.
    } finally {
      setIsExporting(false);
    }
  };

  const exportHARAfterReview = async () => {
    if (!selectedArtifactItem) return;
    const output = await chooseDiagnosticHARPath(
      defaultArtifactExportName(
        selectedArtifactItem.title || selectedArtifactItem.id,
        selectedArtifactItem.createdAt,
      ).replace(/\.dawg$/i, ".har"),
    );
    if (!output) return;
    await engine.exportDiagnosticsHAR(selectedArtifactItem.path, output);
    addLogLine(`Exported sanitized HAR to ${output}.`);
  };

  const copyCurlAfterReview = async () => {
    if (!selectedArtifactItem || !reviewRequestId) return;
    const command = await engine.copyDiagnosticsCurl(selectedArtifactItem.path, reviewRequestId);
    await navigator.clipboard.writeText(command);
    addLogLine(`Copied reviewed cURL for request ${reviewRequestId}.`);
  };

  const removeDiagnosticsAfterReview = async (request: {
    categories: DiagnosticCategory[];
    bodyRefs: string[];
    parentDirectory: string;
    directoryName: string;
  }) => {
    if (!selectedArtifactItem) return;
    const separator = request.parentDirectory.includes("\\") ? "\\" : "/";
    const outputDir = `${request.parentDirectory.replace(/[\\/]+$/, "")}${separator}${request.directoryName}`;
    try {
      const result = await engine.removeDiagnostics({
        artifact: selectedArtifactItem.path,
        outputDir,
        categories: request.categories,
        bodyRefs: request.bodyRefs,
      });
      addLogLine(`Created reviewed artifact without selected diagnostics: ${result.directory}.`);
    } catch (error) {
      addLogLine(`[ERROR] Could not create reviewed artifact: ${String(error)}`);
      throw error;
    }
  };

  const sendReplayControl = async (
    type: ReplayControlType,
    options: { offsetMs?: number; speed?: ReplaySpeed } = {},
  ) => {
    replayCommandSequence.current += 1;
    await engine.controlReplay({
      id: `desktop-${replayCommandSequence.current}`,
      type,
      ...options,
    });
  };

  const commitSeek = async (offsetMs: number) => {
    const target = Math.max(0, Math.min(replayDuration, offsetMs));
    setSeekPreview(null);
    try {
      await sendReplayControl("seek", { offsetMs: target });
    } catch (error) {
      addLogLine(`[ERROR] Could not seek replay: ${String(error)}`);
    }
  };

  const handleRunReplay = async () => {
    if (!selectedArtifact || isReplaying) return;
    replaySequence.current = 0;
    setReplayReady(false);
    setReplayPlaying(false);
    setReplayCurrentTime(0);
    setReplayDuration(0);
    setReplaySpeed(1);
    setSeekPreview(null);
    setIsReplaying(true);
    addLogLine(`Triggering replay for artifact: ${selectedArtifact}`);
    try {
      const result = await engine.runReplay({ artifact: selectedArtifact });
      addLogLine(`Replay ${result.status} for ${result.artifactId}.`);
      const appLogs = result.outcomes?.appLogs?.trim();
      if (appLogs) {
        for (const line of appLogs.split("\n")) {
          if (line.trim()) addLogLine(`[replay] ${line}`);
        }
      }
      const screenshot = result.outcomes?.screenshots?.[0];
      if (screenshot) {
        addLogLine(`Final screenshot: ${screenshot}`);
      }
    } catch (err) {
      if (String(err).includes("cancelled by user")) {
        addLogLine("Replay stopped by user.");
      } else {
        addLogLine(`[ERROR] Replay failed: ${String(err)}`);
      }
    } finally {
      setIsReplaying(false);
      setReplayReady(false);
      setReplayPlaying(false);
      setSeekPreview(null);
    }
  };

  const handleCancelReplay = async () => {
    if (!isReplaying || isCancelling) return;
    setIsCancelling(true);
    addLogLine("Stopping replay — terminating browser and engine process...");
    try {
      await engine.cancelReplay();
    } catch (err) {
      addLogLine(`[ERROR] Failed to stop replay: ${String(err)}`);
    } finally {
      setIsCancelling(false);
    }
  };

  return (
    <PageShell title="Replay Engine" subtitle="Execute deterministic sandboxed artifact replay">
      <Card>
        <CardHeader>
          <CardTitle
            title="Artifact Selection"
            subtitle="Choose a captured session to replay in the sandbox"
          />
        </CardHeader>

        <div className="flex flex-col gap-4">
          <div className="grid grid-cols-1 sm:grid-cols-[minmax(0,1fr)_12rem] gap-3">
            <Input
              id="artifact-search"
              label="Search artifacts"
              placeholder="Name, URL, digest, or path..."
              value={artifactSearch}
              onChange={(event) => setArtifactSearch(event.target.value)}
              iconLeft={<MagnifyingGlass size={14} />}
            />
            <Select
              id="artifact-origin-filter"
              label="Source"
              value={originFilter}
              onChange={setOriginFilter}
              options={[
                { value: "all", label: `All artifacts (${artifacts.length})` },
                { value: "captured", label: "Captured locally" },
                { value: "imported", label: "Imported archives" },
                { value: "pulled", label: "Registry pulls" },
                { value: "legacy", label: "Legacy/local files" },
              ]}
            />
          </div>

          <div className="flex flex-col sm:flex-row gap-3 sm:items-end">
            <Select
              id="artifact-select"
              label="Captured Artifact"
              value={selectedArtifact}
              onChange={setSelectedArtifact}
              wrapperClassName="flex-1 min-w-0 w-full"
              placeholder={
                visibleArtifacts.length === 0
                  ? "No artifacts match this search"
                  : "Select a .dawg artifact..."
              }
              disabled={visibleArtifacts.length === 0}
              options={visibleArtifacts.map((art) => ({
                value: art.path,
                label: `${art.title || art.id}${art.targetUrl ? ` — ${art.targetUrl}` : ""} (${new Date(art.createdAt).toLocaleString()})`,
              }))}
            />

            <div className="flex gap-2 shrink-0 w-full sm:w-auto">
              <Button
                type="button"
                variant="secondary"
                className="flex-1 sm:flex-none"
                onClick={handleImportArtifact}
                disabled={isReplaying}
                loading={isImporting}
                iconLeft={<UploadSimple size={16} />}
              >
                Import
              </Button>
              <Button
                type="button"
                variant="secondary"
                className="flex-1 sm:flex-none"
                onClick={() => setReviewAction("artifact")}
                disabled={!selectedArtifactItem || isReplaying}
                loading={isExporting}
                iconLeft={<Export size={16} />}
              >
                Export
              </Button>
            </div>

            <div className="shrink-0 w-full sm:w-32">
              {isReplaying ? (
                <Button
                  type="button"
                  variant="danger"
                  className="w-full"
                  onClick={handleCancelReplay}
                  disabled={isCancelling}
                  loading={isCancelling}
                  iconLeft={<StopCircle size={16} />}
                >
                  {isCancelling ? "Stopping..." : "Stop Replay"}
                </Button>
              ) : (
                <Button
                  type="button"
                  variant="primary"
                  className="w-full"
                  onClick={handleRunReplay}
                  disabled={!selectedArtifact}
                  iconLeft={<PlayCircle size={16} />}
                >
                  Run Replay
                </Button>
              )}
            </div>
          </div>
        </div>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle
            title="Replay Player"
            subtitle="Synchronized with the controls in replay Chromium"
          />
          <Badge variant={replayReady ? (replayPlaying ? "success" : "info") : "default"} size="sm" dot={replayReady}>
            {replayReady ? (replayPlaying ? "Playing" : "Paused") : isReplaying ? "Starting" : "Idle"}
          </Badge>
        </CardHeader>
        <div className="flex flex-col gap-4">
          <div className="flex flex-wrap items-center gap-2">
            <Button
              type="button"
              variant="secondary"
              size="sm"
              disabled={!replayReady}
              onClick={() =>
                void sendReplayControl(replayPlaying ? "pause" : "play").catch((error) =>
                  addLogLine(`[ERROR] Could not update replay playback: ${String(error)}`),
                )
              }
            >
              {replayPlaying ? "Pause" : "Play"}
            </Button>
            <Button
              type="button"
              variant="secondary"
              size="sm"
              disabled={!replayReady}
              onClick={() => void commitSeek(replayCurrentTime - 10_000)}
            >
              −10s
            </Button>
            <Button
              type="button"
              variant="secondary"
              size="sm"
              disabled={!replayReady}
              onClick={() => void commitSeek(replayCurrentTime + 10_000)}
            >
              +10s
            </Button>
            <Select
              id="replay-speed"
              label="Speed"
              value={String(replaySpeed)}
              disabled={!replayReady}
              onChange={(value) => {
                const speed = Number(value) as ReplaySpeed;
                void sendReplayControl("setSpeed", { speed }).catch((error) =>
                  addLogLine(`[ERROR] Could not change replay speed: ${String(error)}`),
                );
              }}
              options={REPLAY_SPEEDS.map((speed) => ({ value: String(speed), label: `${speed}×` }))}
              wrapperClassName="w-28"
            />
          </div>
          <div className="flex items-center gap-3">
            <input
              aria-label="Replay timeline"
              type="range"
              min={0}
              max={Math.max(0, replayDuration)}
              step={100}
              value={seekPreview ?? replayCurrentTime}
              disabled={!replayReady || replayDuration <= 0}
              onChange={(event) => setSeekPreview(Number(event.target.value))}
              onPointerUp={(event) => void commitSeek(Number(event.currentTarget.value))}
              onKeyUp={(event) => void commitSeek(Number(event.currentTarget.value))}
              className="min-w-0 flex-1 accent-brand-500"
            />
            <output className="shrink-0 text-xs font-mono text-text-secondary">
              {formatReplayTime(seekPreview ?? replayCurrentTime)} / {formatReplayTime(replayDuration)}
            </output>
          </div>
        </div>
      </Card>

      <ReplayDiagnosticInspector
        manifest={manifest}
        diagnosticEvidence={diagnosticEvidence}
        loading={isLoadingManifest}
        error={manifestError}
        replayTimeMs={replayCurrentTime}
        replayActive={replayReady}
        onExportHAR={() => setReviewAction("har")}
        onRemoveDiagnostics={() => setIsDiagnosticRemovalOpen(true)}
        onSeekReplay={(offsetMs) => {
          void engine
            .seekReplay(offsetMs)
            .then(() => addLogLine(`Seeked interactive replay to ${offsetMs}ms.`))
            .catch((error) => addLogLine(`[ERROR] Could not seek replay: ${String(error)}`));
        }}
        onCopyCurl={(requestId) => {
          setReviewRequestId(requestId);
          setReviewAction("curl");
        }}
      />

      <DiagnosticRemovalModal
        open={isDiagnosticRemovalOpen}
        evidence={diagnosticEvidence}
        defaultDirectoryName={
          selectedArtifactItem
            ? defaultArtifactExportName(
                selectedArtifactItem.title || selectedArtifactItem.id,
                selectedArtifactItem.createdAt,
              ).replace(/\.dawg$/i, "-reviewed")
            : "dawg-reviewed"
        }
        onClose={() => setIsDiagnosticRemovalOpen(false)}
        onChooseParentDirectory={chooseDiagnosticRemovalParentDirectory}
        onConfirm={removeDiagnosticsAfterReview}
      />

      <DiagnosticReviewModal
        open={reviewAction !== null}
        action={reviewAction}
        evidence={diagnosticEvidence}
        onClose={() => {
          setReviewAction(null);
          setReviewRequestId(null);
        }}
        onConfirm={async () => {
          if (reviewAction === "artifact") await exportArtifactAfterReview();
          if (reviewAction === "har") await exportHARAfterReview();
          if (reviewAction === "curl") await copyCurlAfterReview();
        }}
      />

      <Card>
        <CardHeader bordered={false}>
          <div className="flex items-center gap-2">
            <Cube size={18} className="text-text-secondary" />
            <CardTitle title="Sandbox Environment" subtitle="Replay isolation status" />
          </div>
        </CardHeader>

        <EmptyState
          icon={<Info size={32} weight="light" />}
          title={sandboxStatus.title}
          description={sandboxStatus.description}
          action={
            <Button variant="secondary" size="sm" iconLeft={<Flask size={14} />}>
              View Setup Guide
            </Button>
          }
        />
      </Card>

      <div className="flex flex-col gap-4 mt-2">
        <h3 className="text-base font-bold text-text-primary">Execution Logs</h3>
        <LogStreamer />
      </div>
    </PageShell>
  );
};
