import { Archive, CheckCircle, Export, MagnifyingGlass, PlayCircle } from "@phosphor-icons/react";
import { useEngine } from "../context/EngineContext";
import { chooseArtifactExportPath, defaultArtifactExportName } from "../lib/artifactDialogs";
import { Badge } from "./ui/Badge";
import { Button } from "./ui/Button";
import { Card } from "./ui/Card";
import { EmptyState } from "./ui/EmptyState";

export function ArtifactList() {
  const { artifacts, addLogLine, exportArtifact, openInspectModal } = useEngine();

  const handleInspect = (art: Parameters<typeof openInspectModal>[0]) => {
    addLogLine(`Inspecting artifact ID ${art.id} at path ${art.path}...`);
    openInspectModal(art);
  };

  const handleReplay = (id: string) => {
    addLogLine(`Triggering sandboxed replay for artifact ID ${id}...`);
  };

  const handleVerify = (id: string) => {
    addLogLine(`Triggering diff verification for artifact ID ${id}...`);
  };

  const handleExport = async (artifact: (typeof artifacts)[number]) => {
    const output = await chooseArtifactExportPath(
      defaultArtifactExportName(artifact.title || artifact.id, artifact.createdAt),
    );
    if (!output) return;
    await exportArtifact(artifact, output);
  };

  const originLabel: Record<(typeof artifacts)[number]["origin"], string> = {
    captured: "Captured locally",
    imported: "Imported archive",
    pulled: "Pulled from registry",
    legacy: "Legacy/local file",
  };

  if (artifacts.length === 0) {
    return (
      <EmptyState
        icon={<Archive size={32} weight="light" />}
        title="No captured artifacts"
        description="Start a capture session to generate .dawg artifacts."
      />
    );
  }

  return (
    <div className="flex flex-col gap-3">
      {artifacts.map((art) => (
        <Card
          key={art.id}
          noPad
          className="p-4 flex flex-col md:flex-row md:items-center justify-between gap-4 hover:border-brand-300 transition-colors"
        >
          <div className="flex flex-col gap-1 min-w-0">
            <div className="flex flex-wrap items-center gap-2.5">
              <span className="font-semibold text-text-primary text-sm truncate" title={art.title || art.id}>
                {art.title || art.id}
              </span>
              <Badge variant="brand" size="sm">
                .dawg
              </Badge>
              <Badge variant={art.origin === "imported" ? "info" : "default"} size="sm">
                {originLabel[art.origin]}
              </Badge>
            </div>
            <div className="text-[11px] text-text-tertiary font-mono truncate max-w-md" title={art.path}>
              {art.path}
            </div>
            <div className="flex flex-wrap items-center gap-2 text-xs text-text-secondary mt-1">
              {art.targetUrl && (
                <>
                  <span>
                    Target: <strong className="text-text-primary font-medium">{art.targetUrl}</strong>
                  </span>
                  <span className="text-border-strong">•</span>
                </>
              )}
              <span>{new Date(art.createdAt).toLocaleString()}</span>
            </div>
          </div>

          <div className="flex items-center gap-2 shrink-0">
            <Button
              variant="secondary"
              size="sm"
              onClick={() => handleInspect(art)}
              iconLeft={<MagnifyingGlass size={14} />}
            >
              Inspect
            </Button>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => handleExport(art)}
              iconLeft={<Export size={14} />}
            >
              Export
            </Button>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => handleReplay(art.id)}
              iconLeft={<PlayCircle size={14} />}
            >
              Replay
            </Button>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => handleVerify(art.id)}
              iconLeft={<CheckCircle size={14} />}
            >
              Verify
            </Button>
          </div>
        </Card>
      ))}
    </div>
  );
}
