import { useEngine } from "../context/EngineContext";

export function ArtifactList() {
  const { artifacts, addLogLine, openInspectModal } = useEngine();

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

  if (artifacts.length === 0) {
    return (
      <div className="bg-slate-900/30 border border-dashed border-slate-700 rounded-md p-8 text-center text-slate-400 text-sm">
        <p>
          No captured artifacts found. Start a capture session to generate .dawg
          artifacts.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {artifacts.map((art) => (
        <div
          key={art.id}
          className="bg-slate-900/60 border border-slate-700 hover:border-slate-600 rounded-lg p-4 flex flex-col md:flex-row md:items-center justify-between gap-4 transition-all"
        >
          <div className="flex flex-col gap-1">
            <div className="flex items-center gap-2">
              <span className="font-semibold text-slate-200 text-sm">
                {art.id}
              </span>
              <span className="text-xs px-2 py-0.5 rounded bg-sky-500/20 text-sky-300 font-mono">
                .dawg v0.1.1-alpha
              </span>
            </div>
            <div className="text-xs text-slate-400 font-mono truncate max-w-md">
              {art.path}
            </div>
            <div className="flex items-center gap-2 text-xs text-slate-400 mt-1">
              <span>
                Target:{" "}
                <strong className="text-slate-300">{art.targetUrl}</strong>
              </span>
              <span>•</span>
              <span>{new Date(art.createdAt).toLocaleString()}</span>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => handleInspect(art)}
              className="px-3 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 rounded text-xs font-medium transition-colors"
            >
              Inspect
            </button>
            <button
              type="button"
              onClick={() => handleReplay(art.id)}
              className="px-3 py-1.5 bg-sky-500/20 hover:bg-sky-500/30 text-sky-300 border border-sky-500/40 rounded text-xs font-semibold transition-colors"
            >
              Replay
            </button>
            <button
              type="button"
              onClick={() => handleVerify(art.id)}
              className="px-3 py-1.5 bg-emerald-500/20 hover:bg-emerald-500/30 text-emerald-300 border border-emerald-500/40 rounded text-xs font-semibold transition-colors"
            >
              Verify
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}
