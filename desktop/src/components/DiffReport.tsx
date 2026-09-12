import { ChartBar, CheckCircle, GitBranch } from "@phosphor-icons/react";
import type React from "react";
import { useState } from "react";
import { useEngine } from "../context/EngineContext";
import { LogStreamer } from "./LogStreamer";
import { Button } from "./ui/Button";
import { Card, CardHeader, CardTitle } from "./ui/Card";
import { EmptyState } from "./ui/EmptyState";
import { Input } from "./ui/Input";
import { PageShell } from "./ui/PageShell";
import { Select } from "./ui/Select";

export const DiffReport: React.FC = () => {
  const { artifacts, addLogLine } = useEngine();
  const [selectedArtifact, setSelectedArtifact] = useState("");
  const [targetBranch, setTargetBranch] = useState("main");

  const handleRunVerify = () => {
    if (!selectedArtifact) return;
    addLogLine(
      `Running verify diff for artifact ${selectedArtifact} against branch ${targetBranch}...`,
    );
  };

  return (
    <PageShell
      title="Diff & Verify"
      subtitle="Compare replayed outcome against target branch or baseline assertions"
    >
      <Card>
        <CardHeader>
          <CardTitle
            title="Verification Setup"
            subtitle="Select artifact and target branch to compare against"
          />
        </CardHeader>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 items-end">
          <Select
            id="verify-artifact-select"
            label="Select Artifact"
            value={selectedArtifact}
            onChange={setSelectedArtifact}
            placeholder="Select a captured .dawg artifact..."
            options={artifacts.map((art) => ({
              value: art.path,
              label: `${art.title || art.id}${art.targetUrl ? ` — ${art.targetUrl}` : ""}`,
            }))}
          />

          <div className="flex gap-2 items-end">
            <Input
              id="target-branch"
              label="Target Branch / Baseline"
              value={targetBranch}
              onChange={(e) => setTargetBranch(e.target.value)}
              wrapperClassName="flex-1"
              iconLeft={<GitBranch size={14} />}
            />
            <Button
              type="button"
              variant="primary"
              onClick={handleRunVerify}
              disabled={!selectedArtifact}
              iconLeft={<CheckCircle size={16} />}
            >
              Verify
            </Button>
          </div>
        </div>
      </Card>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="h-full flex flex-col">
          <CardHeader bordered={false}>
            <div className="flex items-center gap-2">
              <ChartBar size={18} className="text-text-secondary" />
              <CardTitle
                title="Verification Report"
                subtitle="Diff results will appear here after verification"
              />
            </div>
          </CardHeader>

          <div className="flex-1 flex items-center justify-center min-h-[288px]">
            <EmptyState
              icon={<ChartBar size={32} weight="light" />}
              title="No Report Generated"
              description="Select an artifact and run verification to see the diff results."
            />
          </div>
        </Card>

        <div className="flex flex-col h-full">
          <LogStreamer />
        </div>
      </div>
    </PageShell>
  );
};
