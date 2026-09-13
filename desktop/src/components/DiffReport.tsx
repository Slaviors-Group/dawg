import { TerminalWindow } from "@phosphor-icons/react";
import type React from "react";
import { LogStreamer } from "./LogStreamer";
import { Card } from "./ui/Card";
import { EmptyState } from "./ui/EmptyState";
import { PageShell } from "./ui/PageShell";

export const DiffReport: React.FC = () => {
  return (
    <PageShell
      title="Diff & Verify"
      subtitle="Verification is currently available through the engine CLI"
    >
      <Card>
        <EmptyState
          icon={<TerminalWindow size={32} weight="light" />}
          title="Run verification from a terminal"
          description="Use dawg verify <artifact-directory> --against <report-label>. The report label does not select or check out a branch."
        />
      </Card>
      <LogStreamer />
    </PageShell>
  );
};
