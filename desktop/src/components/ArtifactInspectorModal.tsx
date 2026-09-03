import type React from "react";
import type { ArtifactItem } from "../context/EngineContext";

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

  const copyManifest = () => {
    navigator.clipboard.writeText(JSON.stringify(sampleManifest, null, 2));
  };

  return (
    <div className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="bg-slate-900 border border-slate-700 rounded-xl w-full max-w-2xl max-h-[85vh] flex flex-col shadow-2xl overflow-hidden">
        <div className="flex items-center justify-between px-6 py-4 bg-slate-800/80 border-b border-slate-700">
          <div>
            <h3 className="font-semibold text-slate-100 text-lg">Artifact Inspector</h3>
            <p className="text-xs text-slate-400 font-mono">{artifact.id}</p>
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={copyManifest}
              className="px-3 py-1.5 bg-slate-800 hover:bg-slate-700 border border-slate-700 rounded text-xs text-slate-200 transition-colors"
            >
              Copy JSON
            </button>
            <button
              type="button"
              onClick={onClose}
              className="p-1.5 text-slate-400 hover:text-slate-100 hover:bg-slate-800 rounded transition-colors text-lg leading-none"
            >
              &times;
            </button>
          </div>
        </div>

        <div className="p-6 overflow-y-auto space-y-4 font-sans">
          <div className="grid grid-cols-2 gap-3 text-xs bg-slate-950/50 p-3 rounded-lg border border-slate-800">
            <div>
              <span className="text-slate-500 block uppercase tracking-wider text-[10px]">
                Path
              </span>
              <span className="text-slate-200 font-mono break-all">{artifact.path}</span>
            </div>
            <div>
              <span className="text-slate-500 block uppercase tracking-wider text-[10px]">
                Target URL
              </span>
              <span className="text-sky-400 font-medium">{artifact.targetUrl}</span>
            </div>
            <div>
              <span className="text-slate-500 block uppercase tracking-wider text-[10px]">
                Created At
              </span>
              <span className="text-slate-300">
                {new Date(artifact.createdAt).toLocaleString()}
              </span>
            </div>
            <div>
              <span className="text-slate-500 block uppercase tracking-wider text-[10px]">
                Components
              </span>
              <div className="flex gap-1 flex-wrap mt-0.5">
                {artifact.components.map((comp) => (
                  <span
                    key={comp}
                    className="px-1.5 py-0.5 bg-slate-800 text-slate-300 rounded text-[10px] font-mono"
                  >
                    {comp}
                  </span>
                ))}
              </div>
            </div>
          </div>

          <div>
            <h4 className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2">
              Manifest JSON (.dawg/meta.json)
            </h4>
            <pre className="bg-slate-950 border border-slate-800 p-4 rounded-lg font-mono text-xs text-emerald-300 overflow-x-auto max-h-72">
              {JSON.stringify(sampleManifest, null, 2)}
            </pre>
          </div>
        </div>

        <div className="px-6 py-3 bg-slate-800/50 border-t border-slate-800 flex justify-end">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 rounded-md text-sm font-medium transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
};
