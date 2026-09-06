import type React from "react";
import type { ArtifactItem } from "../context/EngineContext";
import { Modal } from "./ui/Modal";
import { Badge } from "./ui/Badge";
import { CodeBlock } from "./ui/CodeBlock";

interface Props {
  artifact: ArtifactItem | null;
  onClose: () => void;
}

export const ArtifactInspectorModal: React.FC<Props> = ({ artifact, onClose }) => {
  if (!artifact) return null;

  const sampleManifest = {
    sessionId: artifact.id,
    startedAt: artifact.createdAt,
    stoppedAt: new Date(new Date(artifact.createdAt).getTime() + 5 * 60000).toISOString(),
    targetUrl: artifact.targetUrl,
    capturedComponents: artifact.components,
    environment: {
      repoCommit: "a1b2c3d",
      branch: "main",
      nodeVersion: "22.14.0",
      goVersion: "1.25.1",
      dockerComposeFile: "env/compose.yaml",
      imageDigests: {
        backend: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
        postgres: "sha256:7d85572974441f14ae28cf856b3e6e5493a952e7d04e515487af920b2d35ef26",
      },
    },
    determinism: {
      clockFrozenAt: artifact.createdAt,
      randomSeed: 42,
    },
    layers: [
      { mediaType: "application/vnd.dawg.env.v1+zstd", size: 1024 },
      { mediaType: "application/vnd.dawg.trace.v1+zstd", size: 45092 },
      { mediaType: "application/vnd.dawg.http.v1+zstd", size: 12048 },
      { mediaType: "application/vnd.dawg.dbdiff.v1+zstd", size: 3020 },
      { mediaType: "application/vnd.dawg.logs.v1+zstd", size: 8192 },
    ],
  };

  return (
    <Modal
      open={!!artifact}
      onClose={onClose}
      title="Artifact Inspector"
      subtitle={artifact.id}
      maxWidth="lg"
    >
      <div className="flex flex-col gap-5">
        {/* Metadata Grid */}
        <div className="grid grid-cols-2 gap-3 p-4 bg-canvas-subtle rounded-md border border-border text-xs">
          <div>
            <span className="block text-[10px] font-medium text-text-tertiary uppercase tracking-wider mb-1">
              Path
            </span>
            <span className="font-mono text-text-primary break-all">
              {artifact.path}
            </span>
          </div>
          <div>
            <span className="block text-[10px] font-medium text-text-tertiary uppercase tracking-wider mb-1">
              Target URL
            </span>
            <span className="font-medium text-brand-600">
              {artifact.targetUrl}
            </span>
          </div>
          <div>
            <span className="block text-[10px] font-medium text-text-tertiary uppercase tracking-wider mb-1">
              Created At
            </span>
            <span className="text-text-secondary">
              {new Date(artifact.createdAt).toLocaleString()}
            </span>
          </div>
          <div>
            <span className="block text-[10px] font-medium text-text-tertiary uppercase tracking-wider mb-1">
              Components
            </span>
            <div className="flex flex-wrap gap-1 mt-0.5">
              {artifact.components.map((comp) => (
                <Badge key={comp} size="sm">{comp}</Badge>
              ))}
            </div>
          </div>
        </div>

        {/* JSON Manifest */}
        <div>
          <h4 className="text-xs font-semibold text-text-secondary mb-2">
            Manifest JSON (.dawg/meta.json)
          </h4>
          <CodeBlock
            code={JSON.stringify(sampleManifest, null, 2)}
            language="json"
            maxHeight="24rem"
          />
        </div>
      </div>
    </Modal>
  );
};
