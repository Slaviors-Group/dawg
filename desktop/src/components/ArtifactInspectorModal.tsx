import type React from "react";
import { useEffect, useState } from "react";
import type { ArtifactItem } from "../context/EngineContext";
import { type InspectResult, engine } from "../lib/engine";
import { Badge } from "./ui/Badge";
import { CodeBlock } from "./ui/CodeBlock";
import { Modal } from "./ui/Modal";

interface Props {
  artifact: ArtifactItem | null;
  onClose: () => void;
}

export const ArtifactInspectorModal: React.FC<Props> = ({ artifact, onClose }) => {
  const [manifest, setManifest] = useState<InspectResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    let disposed = false;
    if (!artifact) {
      setManifest(null);
      setError(null);
      return () => {
        disposed = true;
      };
    }

    setLoading(true);
    setManifest(null);
    setError(null);
    void engine
      .inspectArtifact({ path: artifact.path })
      .then((result) => {
        if (!disposed) setManifest(result);
      })
      .catch((inspectError) => {
        if (!disposed) setError(String(inspectError));
      })
      .finally(() => {
        if (!disposed) setLoading(false);
      });

    return () => {
      disposed = true;
    };
  }, [artifact]);

  if (!artifact) return null;

  return (
    <Modal
      open
      onClose={onClose}
      title="Artifact Inspector"
      subtitle={artifact.title || artifact.id}
      maxWidth="lg"
    >
      <div className="flex min-w-0 flex-col gap-5">
        <div className="grid grid-cols-1 gap-3 rounded-md border border-border bg-canvas-subtle p-4 text-xs sm:grid-cols-2">
          <div className="min-w-0">
            <span className="block text-[10px] font-medium text-text-tertiary uppercase tracking-wider mb-1">
              Path
            </span>
            <span className="font-mono text-text-primary break-all">{artifact.path}</span>
          </div>
          <div className="min-w-0">
            <span className="block text-[10px] font-medium text-text-tertiary uppercase tracking-wider mb-1">
              Target URL
            </span>
            <span className="block font-medium text-brand-600 wrap-anywhere">
              {artifact.targetUrl || "Not recorded"}
            </span>
          </div>
          <div className="min-w-0">
            <span className="block text-[10px] font-medium text-text-tertiary uppercase tracking-wider mb-1">
              Created At
            </span>
            <span className="text-text-secondary">
              {new Date(artifact.createdAt).toLocaleString()}
            </span>
          </div>
          <div className="min-w-0">
            <span className="block text-[10px] font-medium text-text-tertiary uppercase tracking-wider mb-1">
              Components
            </span>
            <div className="flex flex-wrap gap-1 mt-0.5">
              {artifact.components.map((component) => (
                <Badge
                  key={component}
                  size="sm"
                  className="max-w-full whitespace-normal wrap-anywhere"
                >
                  {component}
                </Badge>
              ))}
            </div>
          </div>
        </div>

        <div className="min-w-0">
          <h4 className="text-xs font-semibold text-text-secondary mb-2">DAWG manifest</h4>
          {loading && <p className="text-sm text-text-secondary">Loading manifest…</p>}
          {error && <p className="text-sm text-error-text wrap-anywhere">{error}</p>}
          {manifest && (
            <CodeBlock code={JSON.stringify(manifest, null, 2)} language="json" maxHeight="24rem" />
          )}
        </div>
      </div>
    </Modal>
  );
};
