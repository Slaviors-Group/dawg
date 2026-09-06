import type React from "react";
import { useState } from "react";
import { useEngine } from "../context/EngineContext";
import { LogStreamer } from "./LogStreamer";
import { PageShell } from "./ui/PageShell";
import { Card, CardHeader, CardTitle } from "./ui/Card";
import { Select } from "./ui/Select";
import { Button } from "./ui/Button";
import { EmptyState } from "./ui/EmptyState";
import { Separator } from "./ui/Separator";
import {
  PlayCircle,
  WarningCircle,
  Cube,
  Flask,
} from "@phosphor-icons/react";

export const ReplayViewer: React.FC = () => {
  const { artifacts, addLogLine } = useEngine();
  const [selectedArtifact, setSelectedArtifact] = useState("");

  const handleRunReplay = () => {
    if (!selectedArtifact) return;
    addLogLine(`Triggering sandboxed replay for artifact: ${selectedArtifact}`);
  };

  return (
    <PageShell
      title="Replay Engine"
      subtitle="Execute deterministic sandboxed artifact replay"
    >
      <Card>
        <CardHeader>
          <CardTitle
            title="Artifact Selection"
            subtitle="Choose a captured session to replay in the sandbox"
          />
        </CardHeader>

        <div className="flex flex-col gap-4">
          <div className="flex gap-3 items-end">
            <Select
              id="artifact-select"
              label="Captured Artifact"
              value={selectedArtifact}
              onChange={setSelectedArtifact}
              wrapperClassName="flex-1"
              placeholder="Select a .dawg artifact..."
              options={artifacts.map((art) => ({
                value: art.path,
                label: `${art.id} — ${art.targetUrl} (${new Date(art.createdAt).toLocaleTimeString()})`
              }))}
            />

            <Button
              type="button"
              variant="primary"
              onClick={handleRunReplay}
              disabled={!selectedArtifact}
              iconLeft={<PlayCircle size={16} />}
            >
              Run Replay
            </Button>
          </div>
        </div>
      </Card>

      <Card>
        <CardHeader bordered={false}>
          <div className="flex items-center gap-2">
            <Cube size={18} className="text-text-secondary" />
            <CardTitle
              title="Sandbox Environment"
              subtitle="Replay isolation status"
            />
          </div>
        </CardHeader>
        
        <EmptyState
          icon={<WarningCircle size={32} weight="light" />}
          title="Sandbox requires Linux / WSL2"
          description="Bare Windows hosts return sandbox isolation requirement error per security architecture. Rootless Docker sandbox is needed."
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
