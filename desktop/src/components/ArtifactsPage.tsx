import {
  CaretDown,
  CaretRight,
  DownloadSimple,
  MagnifyingGlass,
  PencilSimple,
  PlayCircle,
  Trash,
  UploadSimple,
} from "@phosphor-icons/react";
import { getCurrentWindow } from "@tauri-apps/api/window";

import { useCallback, useEffect, useMemo, useState } from "react";

import { type ArtifactItem, useEngine } from "../context/EngineContext";
import {
  chooseArtifactArchive,
  chooseArtifactExportPath,
  defaultArtifactExportName,
} from "../lib/artifactDialogs";
import { ArtifactInspectorModal } from "./ArtifactInspectorModal";
import { Badge } from "./ui/Badge";
import { Button } from "./ui/Button";
import { Card } from "./ui/Card";
import { EmptyState } from "./ui/EmptyState";
import { Input } from "./ui/Input";
import { Modal } from "./ui/Modal";
import { PageShell } from "./ui/PageShell";
import { Select } from "./ui/Select";

interface ArtifactsPageProps {
  onReplayArtifact: (artifact: ArtifactItem) => void;
  onEditArtifact: (artifact: ArtifactItem) => void;
}

type SortMode = "added" | "created" | "title" | "flags";

const originLabel: Record<ArtifactItem["origin"], string> = {
  captured: "Captured",
  imported: "Imported",
  pulled: "Pulled",
  legacy: "Legacy",
};

const artifactIdentity = (artifact: ArtifactItem) => artifact.instanceId || artifact.path;

const dateValue = (value?: string) => {
  const parsed = value ? new Date(value).getTime() : Number.NaN;
  return Number.isFinite(parsed) ? parsed : 0;
};

export function ArtifactsPage({ onReplayArtifact, onEditArtifact }: ArtifactsPageProps) {
  const {
    artifacts,
    addLogLine,
    importArtifact,
    exportArtifact,
    deleteArtifact,
    inspectedArtifact,
    openInspectModal,
    closeInspectModal,
  } = useEngine();
  const [query, setQuery] = useState("");
  const [origin, setOrigin] = useState("all");
  const [flagged, setFlagged] = useState("all");
  const [sort, setSort] = useState<SortMode>("added");
  const [expandedRoots, setExpandedRoots] = useState<Set<string>>(new Set());
  const [deleteTarget, setDeleteTarget] = useState<ArtifactItem | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  const [isImporting, setIsImporting] = useState(false);
  const [isDragOver, setIsDragOver] = useState(false);

  const importArchive = useCallback(
    async (archive: string) => {
      if (!archive.toLowerCase().endsWith(".dawg")) {
        addLogLine("[WARN] Only .dawg artifact archives can be imported.");
        return;
      }
      setIsImporting(true);
      try {
        await importArtifact(archive);
      } catch {
        // The shared engine log contains the invocation error.
      } finally {
        setIsImporting(false);
      }
    },
    [addLogLine, importArtifact],
  );

  useEffect(() => {
    let disposed = false;
    let unlisten: (() => void) | undefined;
    void getCurrentWindow()
      .onDragDropEvent(({ payload }) => {
        if (payload.type === "enter" || payload.type === "over") setIsDragOver(true);
        if (payload.type === "leave") setIsDragOver(false);
        if (payload.type === "drop") {
          setIsDragOver(false);
          const archive = payload.paths.find((path) => path.toLowerCase().endsWith(".dawg"));
          if (archive) void importArchive(archive);
          else addLogLine("[WARN] Drop a .dawg artifact archive to import it.");
        }
      })
      .then((stopListening) => {
        if (disposed) stopListening();
        else unlisten = stopListening;
      })
      .catch((error) => addLogLine(`[WARN] Drag-and-drop import is unavailable: ${String(error)}`));
    return () => {
      disposed = true;
      unlisten?.();
    };
  }, [addLogLine, importArchive]);

  const groupedArtifacts = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();
    const filtered = artifacts.filter((artifact) => {
      if (origin !== "all" && artifact.origin !== origin) return false;
      if (flagged === "flagged" && !artifact.flagged) return false;
      if (flagged === "unflagged" && artifact.flagged) return false;
      if (!normalizedQuery) return true;
      return [
        artifact.title,
        artifact.id,
        artifact.targetUrl,
        artifact.path,
        ...(artifact.flagTitles ?? []),
      ]
        .join(" ")
        .toLowerCase()
        .includes(normalizedQuery);
    });

    const byRoot = new Map<string, ArtifactItem[]>();
    for (const artifact of filtered) {
      const root = artifact.reviewRootId || artifact.rootArtifactId || artifact.id;
      byRoot.set(root, [...(byRoot.get(root) ?? []), artifact]);
    }

    const compare = (left: ArtifactItem, right: ArtifactItem) => {
      if (sort === "title") return (left.title || left.id).localeCompare(right.title || right.id);
      if (sort === "flags") return (right.flagCount ?? 0) - (left.flagCount ?? 0);
      const leftDate =
        sort === "created" ? dateValue(left.createdAt) : dateValue(left.addedAt || left.createdAt);
      const rightDate =
        sort === "created"
          ? dateValue(right.createdAt)
          : dateValue(right.addedAt || right.createdAt);
      return rightDate - leftDate;
    };

    return [...byRoot.entries()]
      .map(([root, revisions]) => ({
        root,
        revisions: [...revisions].sort((left, right) => {
          const revisionDifference = (right.revision ?? 0) - (left.revision ?? 0);
          return revisionDifference || compare(left, right);
        }),
      }))
      .sort((left, right) => compare(left.revisions[0], right.revisions[0]));
  }, [artifacts, flagged, origin, query, sort]);

  const handleImport = async () => {
    const archive = await chooseArtifactArchive();
    if (archive) await importArchive(archive);
  };

  const handleExport = async (artifact: ArtifactItem) => {
    const output = await chooseArtifactExportPath(
      defaultArtifactExportName(artifact.title || artifact.id, artifact.createdAt),
    );
    if (!output) return;
    try {
      await exportArtifact(artifact, output);
    } catch {
      // The shared engine log contains the invocation error.
    }
  };

  const confirmDelete = async () => {
    if (!deleteTarget) return;
    setIsDeleting(true);
    try {
      await deleteArtifact(deleteTarget);
      setDeleteTarget(null);
    } catch {
      // Keep the confirmation open so the engine's protective error can be reviewed.
    } finally {
      setIsDeleting(false);
    }
  };

  const toggleRoot = (root: string) => {
    setExpandedRoots((current) => {
      const next = new Set(current);
      if (next.has(root)) next.delete(root);
      else next.add(root);
      return next;
    });
  };

  return (
    <PageShell
      title="Artifacts"
      subtitle="Search, review, export, and safely manage local artifact instances"
      className="h-full min-h-0 overflow-hidden"
      bodyClassName="min-h-0 flex-1 overflow-hidden"
    >
      <Card className="shrink-0">
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-3 lg:flex-row lg:items-end">
            <Input
              id="artifact-catalog-search"
              label="Search catalog"
              placeholder="Title, ID, URL, path, or review flag..."
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              iconLeft={<MagnifyingGlass size={15} />}
              wrapperClassName="flex-1"
            />
            <Select
              id="artifact-catalog-origin"
              label="Origin"
              value={origin}
              onChange={setOrigin}
              options={[
                { value: "all", label: "All origins" },
                { value: "captured", label: "Captured" },
                { value: "imported", label: "Imported" },
                { value: "pulled", label: "Pulled" },
                { value: "legacy", label: "Legacy" },
              ]}
              wrapperClassName="lg:w-36"
            />
            <Select
              id="artifact-catalog-flagged"
              label="Review"
              value={flagged}
              onChange={setFlagged}
              options={[
                { value: "all", label: "All artifacts" },
                { value: "flagged", label: "Flagged" },
                { value: "unflagged", label: "Unflagged" },
              ]}
              wrapperClassName="lg:w-36"
            />
            <Select
              id="artifact-catalog-sort"
              label="Sort"
              value={sort}
              onChange={(value) => setSort(value as SortMode)}
              options={[
                { value: "added", label: "Recently added" },
                { value: "created", label: "Capture date" },
                { value: "title", label: "Title" },
                { value: "flags", label: "Flag count" },
              ]}
              wrapperClassName="lg:w-40"
            />
            <Button
              variant="secondary"
              onClick={handleImport}
              loading={isImporting}
              iconLeft={<UploadSimple size={16} />}
            >
              Import
            </Button>
          </div>
          <div
            className={[
              "rounded-lg border-2 border-dashed px-4 py-3 text-center transition-colors",
              isDragOver
                ? "border-brand-500 bg-brand-100/50 text-brand-700"
                : "border-border bg-canvas-subtle text-text-secondary",
            ].join(" ")}
          >
            <p className="text-xs font-medium">
              Drop a .dawg archive anywhere in this catalog to import it
            </p>
          </div>
        </div>
      </Card>

      <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain pr-2">
        {groupedArtifacts.length === 0 ? (
          <EmptyState
            icon={<DownloadSimple size={32} weight="light" />}
            title={
              artifacts.length === 0 ? "No artifact instances" : "No artifacts match these filters"
            }
            description={
              artifacts.length === 0
                ? "Capture a session or import a .dawg archive to build the local catalog."
                : "Try clearing a search term or changing the catalog filters."
            }
          />
        ) : (
          <div className="flex flex-col gap-3">
            {groupedArtifacts.map(({ root, revisions }) => {
              const latest = revisions[0];
              const expanded = expandedRoots.has(root);
              return (
                <div key={root} className="flex flex-col gap-2">
                  <ArtifactRow
                    artifact={latest}
                    revisionCount={revisions.length}
                    onInspect={openInspectModal}
                    onReplay={onReplayArtifact}
                    onEdit={onEditArtifact}
                    onExport={handleExport}
                    onDelete={setDeleteTarget}
                    onToggleRevisions={revisions.length > 1 ? () => toggleRoot(root) : undefined}
                    revisionsExpanded={expanded}
                  />
                  {expanded &&
                    revisions.slice(1).map((artifact) => (
                      <div
                        key={artifactIdentity(artifact)}
                        className="ml-4 border-l-2 border-brand-200 pl-3"
                      >
                        <ArtifactRow
                          artifact={artifact}
                          onInspect={openInspectModal}
                          onReplay={onReplayArtifact}
                          onEdit={onEditArtifact}
                          onExport={handleExport}
                          onDelete={setDeleteTarget}
                        />
                      </div>
                    ))}
                </div>
              );
            })}
          </div>
        )}
      </div>

      <ArtifactInspectorModal artifact={inspectedArtifact} onClose={closeInspectModal} />
      <Modal
        open={Boolean(deleteTarget)}
        onClose={() => !isDeleting && setDeleteTarget(null)}
        title="Delete local artifact instance?"
        subtitle={deleteTarget?.title || deleteTarget?.id}
        maxWidth="sm"
        footer={
          <div className="flex justify-end gap-2">
            <Button
              variant="secondary"
              size="sm"
              disabled={isDeleting}
              onClick={() => setDeleteTarget(null)}
            >
              Cancel
            </Button>
            <Button
              variant="danger"
              size="sm"
              loading={isDeleting}
              onClick={() => void confirmDelete()}
            >
              Delete instance
            </Button>
          </div>
        }
      >
        <div className="flex flex-col gap-3 text-sm text-text-secondary">
          <p>
            This removes only this managed local instance. The engine rejects artifacts that are
            active, outside its managed store, or otherwise unsafe to delete.
          </p>
          {deleteTarget && (
            <p className="rounded-md border border-border bg-canvas-subtle p-3 font-mono text-xs break-all">
              {deleteTarget.path}
            </p>
          )}
        </div>
      </Modal>
    </PageShell>
  );
}

interface ArtifactRowProps {
  artifact: ArtifactItem;
  revisionCount?: number;
  revisionsExpanded?: boolean;
  onInspect: (artifact: ArtifactItem) => void;
  onReplay: (artifact: ArtifactItem) => void;
  onEdit: (artifact: ArtifactItem) => void;
  onExport: (artifact: ArtifactItem) => Promise<void>;
  onDelete: (artifact: ArtifactItem) => void;
  onToggleRevisions?: () => void;
}

function ArtifactRow({
  artifact,
  revisionCount,
  revisionsExpanded,
  onInspect,
  onReplay,
  onEdit,
  onExport,
  onDelete,
  onToggleRevisions,
}: ArtifactRowProps) {
  return (
    <Card
      noPad
      className="flex flex-col gap-4 p-4 transition-colors hover:border-brand-300 xl:flex-row xl:items-center xl:justify-between"
    >
      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-2">
          <span
            className="truncate text-sm font-semibold text-text-primary"
            title={artifact.title || artifact.id}
          >
            {artifact.title || artifact.id}
          </span>
          <Badge variant="brand" size="sm">
            .dawg
          </Badge>
          <Badge variant={artifact.origin === "imported" ? "info" : "default"} size="sm">
            {originLabel[artifact.origin]}
          </Badge>
          {artifact.flagged && (
            <Badge variant="warning" size="sm">
              Flagged{artifact.flagCount ? ` · ${artifact.flagCount}` : ""}
            </Badge>
          )}
          {artifact.revision && (
            <Badge variant="info" size="sm">
              Revision {artifact.revision}
            </Badge>
          )}
          {artifact.tags?.map((tag) => (
            <Badge key={tag} size="sm">
              {tag}
            </Badge>
          ))}
          {revisionCount && revisionCount > 1 && (
            <button
              type="button"
              onClick={onToggleRevisions}
              className="inline-flex items-center gap-1 text-xs font-medium text-brand-600 hover:text-brand-700"
              aria-expanded={revisionsExpanded}
            >
              {revisionsExpanded ? <CaretDown size={13} /> : <CaretRight size={13} />}
              {revisionCount} revisions
            </button>
          )}
        </div>
        <p className="mt-1 truncate font-mono text-[11px] text-text-tertiary" title={artifact.path}>
          {artifact.path}
        </p>
        <div className="mt-2 flex flex-wrap gap-x-2 gap-y-1 text-xs text-text-secondary">
          {artifact.targetUrl && <span className="truncate">{artifact.targetUrl}</span>}
          {artifact.targetUrl && <span className="text-border-strong">•</span>}
          <span>Added {new Date(artifact.addedAt || artifact.createdAt).toLocaleString()}</span>
          {artifact.instanceId && (
            <>
              <span className="text-border-strong">•</span>
              <span className="font-mono">{artifact.instanceId}</span>
            </>
          )}
        </div>
      </div>
      <div className="flex flex-wrap gap-2">
        <Button
          variant="secondary"
          size="sm"
          onClick={() => onInspect(artifact)}
          iconLeft={<MagnifyingGlass size={14} />}
        >
          Inspect
        </Button>
        <Button
          variant="secondary"
          size="sm"
          onClick={() => onReplay(artifact)}
          iconLeft={<PlayCircle size={14} />}
        >
          Replay
        </Button>
        <Button
          variant="secondary"
          size="sm"
          onClick={() => onEdit(artifact)}
          iconLeft={<PencilSimple size={14} />}
        >
          Edit
        </Button>
        <Button
          variant="secondary"
          size="sm"
          onClick={() => void onExport(artifact)}
          iconLeft={<DownloadSimple size={14} />}
        >
          Export
        </Button>
        <Button
          variant="danger"
          size="sm"
          onClick={() => onDelete(artifact)}
          iconLeft={<Trash size={14} />}
        >
          Delete
        </Button>
      </div>
    </Card>
  );
}
