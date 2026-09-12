import { Cube, Flask, Info, MagnifyingGlass, PlayCircle, StopCircle } from "@phosphor-icons/react";
import type React from "react";
import { useMemo, useState } from "react";
import { useEngine } from "../context/EngineContext";
import { engine } from "../lib/engine";
import { LogStreamer } from "./LogStreamer";
import { Button } from "./ui/Button";
import { Card, CardHeader, CardTitle } from "./ui/Card";
import { EmptyState } from "./ui/EmptyState";
import { Input } from "./ui/Input";
import { PageShell } from "./ui/PageShell";
import { Select } from "./ui/Select";

// The desktop shell has no reliable IPC signal for host OS yet, so this is a
// best-effort UI hint only. The engine itself is the source of truth: it
// already falls back to a native (non-containerized) replay path on bare
// Windows instead of hard-failing (see replay.Sandbox.Start / cmd/dawg/run.go).
const isWindows = typeof navigator !== "undefined" && navigator.userAgent.includes("Windows");

export const ReplayViewer: React.FC = () => {
  const { artifacts, addLogLine } = useEngine();
  const [selectedArtifact, setSelectedArtifact] = useState("");
  const [artifactSearch, setArtifactSearch] = useState("");
  const [originFilter, setOriginFilter] = useState("all");
  const [isReplaying, setIsReplaying] = useState(false);
  const [isCancelling, setIsCancelling] = useState(false);

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

  const handleRunReplay = async () => {
    if (!selectedArtifact || isReplaying) return;
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
