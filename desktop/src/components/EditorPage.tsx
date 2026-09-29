import {
  CheckCircle,
  PencilSimple,
  PlayCircle,
  StopCircle,
  WarningCircle,
} from "@phosphor-icons/react";
import { listen } from "@tauri-apps/api/event";
import type React from "react";
import { useEffect, useMemo, useRef, useState } from "react";

import { type ArtifactItem, useEngine } from "../context/EngineContext";
import { type EditorEvent, type RunEditorResult, engine } from "../lib/engine";
import { Badge } from "./ui/Badge";
import { Button } from "./ui/Button";
import { Card, CardHeader, CardTitle } from "./ui/Card";
import { EmptyState } from "./ui/EmptyState";
import { PageShell } from "./ui/PageShell";
import { Select } from "./ui/Select";

interface EditorPageProps {
  selectedArtifactIdentity?: string;
  onReplayArtifact: (artifactIdentity: string) => void;
}

const artifactIdentity = (artifact: ArtifactItem) => artifact.instanceId || artifact.path;

export const EditorPage: React.FC<EditorPageProps> = ({
  selectedArtifactIdentity,
  onReplayArtifact,
}) => {
  const { artifacts, addLogLine, refreshArtifacts } = useEngine();
  const [selectedIdentity, setSelectedIdentity] = useState("");
  const [isLaunching, setIsLaunching] = useState(false);
  const [isStopping, setIsStopping] = useState(false);
  const [editorState, setEditorState] = useState<
    "idle" | "starting" | "ready" | "dirty" | "saving" | "closing" | "error"
  >("idle");
  const [message, setMessage] = useState("Select an artifact to open its Chromium review editor.");
  const [savedArtifactIdentity, setSavedArtifactIdentity] = useState<string | null>(null);
  const eventSequence = useRef(0);
  const activeEditorRun = useRef<Promise<RunEditorResult> | null>(null);
  const stopRequested = useRef(false);
  const closedEventReceived = useRef(false);

  useEffect(() => {
    if (
      selectedArtifactIdentity &&
      artifacts.some((artifact) => artifactIdentity(artifact) === selectedArtifactIdentity)
    ) {
      setSelectedIdentity(selectedArtifactIdentity);
    }
  }, [artifacts, selectedArtifactIdentity]);

  useEffect(() => {
    let disposed = false;
    let unlisten: (() => void) | undefined;
    void listen<EditorEvent>("dawg://editor-event", ({ payload }) => {
      if (payload.protocol !== "dawg.editor.v1" || payload.sequence <= eventSequence.current)
        return;
      eventSequence.current = payload.sequence;
      if (payload.type === "ready") {
        closedEventReceived.current = false;
        setIsLaunching(false);
        setEditorState("ready");
        setMessage("Chromium Editor is ready. Changes are saved as a local draft until published.");
      }
      if (payload.type === "draftChanged") {
        setEditorState("dirty");
        setMessage("Review draft has unsaved changes.");
      }
      if (payload.type === "draftSaved") {
        setEditorState("ready");
        setMessage("Review draft saved locally.");
      }
      if (payload.type === "validationError" || payload.type === "error") {
        setEditorState("error");
        setMessage(payload.message || "The editor rejected the requested change.");
        addLogLine(`[ERROR] Editor: ${payload.message || "validation error"}`);
      }
      if (payload.type === "artifactSaved") {
        setEditorState("ready");
        const identity = payload.instanceId || payload.artifactPath || payload.artifactId || null;
        setSavedArtifactIdentity(identity);
        setMessage(
          payload.artifactTitle
            ? `Saved “${payload.artifactTitle}” as a new immutable revision.`
            : "A new immutable flagged artifact revision was published.",
        );
        addLogLine(
          `Editor saved artifact ${payload.artifactId || payload.artifactPath || "revision"}.`,
        );
        void refreshArtifacts();
      }
      if (payload.type === "closed") {
        closedEventReceived.current = true;
        setEditorState("closing");
        setMessage(payload.message || "Chromium Editor closed. Finishing cleanup…");
      }
    })
      .then((stopListening) => {
        if (disposed) stopListening();
        else unlisten = stopListening;
      })
      .catch((error) =>
        addLogLine(`[WARN] Editor status events are unavailable: ${String(error)}`),
      );
    return () => {
      disposed = true;
      unlisten?.();
    };
  }, [addLogLine, refreshArtifacts]);

  const selectedArtifact = useMemo(
    () => artifacts.find((artifact) => artifactIdentity(artifact) === selectedIdentity),
    [artifacts, selectedIdentity],
  );

  const savedArtifact = useMemo(
    () =>
      savedArtifactIdentity
        ? artifacts.find(
            (artifact) =>
              artifactIdentity(artifact) === savedArtifactIdentity ||
              artifact.path === savedArtifactIdentity ||
              artifact.id === savedArtifactIdentity,
          )
        : undefined,
    [artifacts, savedArtifactIdentity],
  );

  const launchEditor = async () => {
    if (!selectedArtifact || isLaunching || isStopping || activeEditorRun.current) return;
    eventSequence.current = 0;
    stopRequested.current = false;
    closedEventReceived.current = false;
    setIsLaunching(true);
    setEditorState("starting");
    setMessage("Launching Chromium Editor…");
    addLogLine(`Launching Editor for ${selectedArtifact.id} at ${selectedArtifact.path}.`);
    const run = engine.launchEditor({ artifact: selectedArtifact.path });
    activeEditorRun.current = run;
    try {
      const result = await run;
      if (result.status === "completed") {
        setEditorState("idle");
        setMessage(
          closedEventReceived.current ? "Chromium Editor closed." : "Chromium Editor exited.",
        );
      }
    } catch (error) {
      if (stopRequested.current) {
        setEditorState("idle");
        setMessage("Chromium Editor stopped.");
      } else {
        setEditorState("error");
        setMessage(String(error));
        addLogLine(`[ERROR] Editor launch failed: ${String(error)}`);
      }
    } finally {
      if (activeEditorRun.current === run) activeEditorRun.current = null;
      setIsLaunching(false);
    }
  };

  const stopEditor = async () => {
    setIsStopping(true);
    stopRequested.current = true;
    try {
      const cancelled = await engine.cancelEditor();
      if (!cancelled) {
        stopRequested.current = false;
        setEditorState("idle");
        setMessage("No active Chromium Editor process was found.");
        return;
      }
      setMessage("Stopping Chromium Editor…");
      await activeEditorRun.current?.catch(() => undefined);
      setEditorState("idle");
      setMessage("Chromium Editor stopped.");
    } catch (error) {
      stopRequested.current = false;
      setEditorState("error");
      setMessage(String(error));
      addLogLine(`[ERROR] Could not stop Editor: ${String(error)}`);
    } finally {
      setIsStopping(false);
    }
  };

  const statusVariant =
    editorState === "error"
      ? "error"
      : editorState === "ready" || editorState === "dirty"
        ? "success"
        : "default";
  const statusLabel =
    editorState === "dirty" ? "Draft changed" : editorState[0].toUpperCase() + editorState.slice(1);

  return (
    <PageShell
      title="Editor"
      subtitle="Launch Chromium review editing; the engine owns drafts and immutable artifact publication"
    >
      <Card>
        <CardHeader>
          <CardTitle
            title="Artifact selection"
            subtitle="Select the local artifact instance to review"
          />
          <Badge
            variant={statusVariant}
            size="sm"
            dot={editorState === "ready" || editorState === "dirty"}
          >
            {statusLabel}
          </Badge>
        </CardHeader>
        <div className="flex flex-col gap-4">
          <Select
            id="editor-artifact-select"
            label="Artifact"
            value={selectedIdentity}
            onChange={setSelectedIdentity}
            placeholder={artifacts.length ? "Select an artifact…" : "No local artifacts available"}
            disabled={artifacts.length === 0 || isLaunching || isStopping}
            options={artifacts.map((artifact) => ({
              value: artifactIdentity(artifact),
              label: `${artifact.title || artifact.id}${artifact.flagged ? ` · ${artifact.flagCount ?? 0} flags` : ""}`,
            }))}
          />
          {selectedArtifact && (
            <div className="grid grid-cols-1 gap-3 rounded-md border border-border bg-canvas-subtle p-4 text-xs sm:grid-cols-3">
              <div>
                <span className="block text-text-tertiary">Review flags</span>
                <strong className="text-text-primary">{selectedArtifact.flagCount ?? 0}</strong>
              </div>
              <div>
                <span className="block text-text-tertiary">Revision</span>
                <strong className="text-text-primary">
                  {selectedArtifact.revision ?? "Original"}
                </strong>
              </div>
              <div>
                <span className="block text-text-tertiary">Origin</span>
                <strong className="text-text-primary capitalize">{selectedArtifact.origin}</strong>
              </div>
            </div>
          )}
          <div className="flex flex-wrap gap-2">
            {editorState === "starting" ||
            editorState === "ready" ||
            editorState === "dirty" ||
            editorState === "saving" ? (
              <Button
                variant="danger"
                loading={isStopping}
                onClick={() => void stopEditor()}
                iconLeft={<StopCircle size={16} />}
              >
                Stop Editor
              </Button>
            ) : (
              <Button
                disabled={
                  !selectedArtifact ||
                  isStopping ||
                  editorState === "closing" ||
                  activeEditorRun.current !== null
                }
                loading={isLaunching}
                onClick={() => void launchEditor()}
                iconLeft={<PencilSimple size={16} />}
              >
                {editorState === "closing"
                  ? "Finishing cleanup…"
                  : editorState === "idle"
                    ? "Launch Editor"
                    : "Reopen Editor"}
              </Button>
            )}
            {savedArtifact && (
              <Button
                variant="secondary"
                onClick={() => onReplayArtifact(artifactIdentity(savedArtifact))}
                iconLeft={<PlayCircle size={16} />}
              >
                Replay saved revision
              </Button>
            )}
          </div>
        </div>
      </Card>

      <Card>
        <div className="flex items-start gap-3">
          {editorState === "error" ? (
            <WarningCircle size={20} className="shrink-0 text-error-text" />
          ) : (
            <CheckCircle size={20} className="shrink-0 text-brand-500" />
          )}
          <div className="min-w-0">
            <h3 className="text-sm font-semibold text-text-primary">Editor status</h3>
            <p className="mt-1 text-sm text-text-secondary wrap-anywhere">{message}</p>
            <p className="mt-2 text-xs text-text-tertiary">
              Saving in Chromium creates a new artifact revision; it never changes the source
              artifact. You can choose its name in Chromium, or leave it blank to use the source
              title plus the revision number. Validation and active-process errors are shown here
              and in the shared engine log.
            </p>
          </div>
        </div>
      </Card>

      {artifacts.length === 0 && (
        <EmptyState
          icon={<PencilSimple size={32} weight="light" />}
          title="Nothing to edit yet"
          description="Capture or import a .dawg artifact, then return here to add portable review flags."
        />
      )}
    </PageShell>
  );
};
