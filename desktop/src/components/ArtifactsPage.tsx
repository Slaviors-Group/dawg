import {
  Archive,
  CaretDown,
  CaretRight,
  DownloadSimple,
  Globe,
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

const targetLabel = (value: string) => {
  try {
    const url = new URL(value);
    return `${url.host}${url.pathname === "/" ? "" : url.pathname}`;
  } catch {
    return value;
  }
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
        artifact.instanceId,
        artifact.targetUrl,
        artifact.path,
        ...(artifact.tags ?? []),
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
  const matchCount = groupedArtifacts.reduce((count, group) => count + group.revisions.length, 0);

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
      subtitle="Search, review, and export local artifacts."
      className="h-full min-h-0 overflow-hidden"
      bodyClassName="min-h-0 flex-1 overflow-hidden"
    >
      <Card className="shrink-0">
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-3 lg:flex-row lg:items-end">
            <Input
              data-tour="artifacts-search"
              id="artifact-catalog-search"
              label="Search catalog"
              placeholder="Search name, URL, ID, or path..."
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              onKeyDown={(event) => {
                if (event.key !== "Enter" || !query.trim()) return;
                event.preventDefault();
                const firstMatch = groupedArtifacts[0]?.revisions[0];
                if (firstMatch) openInspectModal(firstMatch);
              }}
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
              data-tour="artifacts-import"
              variant="secondary"
              onClick={handleImport}
              loading={isImporting}
              iconLeft={<UploadSimple size={16} />}
            >
              Import
            </Button>
          </div>
          {query.trim() && (
            <output className="text-xs text-text-tertiary">
              {matchCount} {matchCount === 1 ? "match" : "matches"}. Enter to inspect the first.
            </output>
          )}
          <div
            className={[
              "rounded-2xl border border-dashed px-4 py-3 text-center transition-colors",
              isDragOver
                ? "border-brand-500 bg-brand-100/50 text-brand-700"
                : "border-border bg-canvas-subtle text-text-secondary",
            ].join(" ")}
          >
            <p className="text-xs font-medium">Drop a .dawg archive here to import it</p>
          </div>
        </div>
      </Card>

      <div
        data-tour="artifacts-list"
        className="min-h-0 flex-1 overflow-y-auto overscroll-contain pr-2"
      >
        {groupedArtifacts.length === 0 ? (
          <EmptyState
            icon={<DownloadSimple size={32} weight="light" />}
            title={
              artifacts.length === 0 ? "No artifact instances" : "No artifacts match these filters"
            }
            description={
              artifacts.length === 0
                ? "Capture a session or import a DAWG archive."
                : "Clear the search or change the filters."
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
                      <div key={artifactIdentity(artifact)} className="ml-4">
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
        title="Delete artifact?"
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
          <p>This deletes this local instance. Active or unmanaged artifacts cannot be deleted.</p>
          {deleteTarget && (
            <p className="break-all rounded-xl border border-border bg-canvas-subtle p-3 font-mono text-xs">
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
  const addedAt = new Date(artifact.addedAt || artifact.createdAt);
  const addedLabel = Number.isNaN(addedAt.getTime())
    ? "Date unavailable"
    : `Added ${addedAt.toLocaleDateString(undefined, {
        month: "short",
        day: "numeric",
        year: "numeric",
      })}`;

  return (
    <Card noPad className="p-4 transition-colors hover:border-brand-300 sm:p-5">
      <div className="flex min-w-0 items-start gap-3 sm:gap-4">
        <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-brand-100 text-brand-600">
          <Archive size={20} weight="duotone" aria-hidden />
        </span>
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-start justify-between gap-2">
            <h3
              className="min-w-0 truncate text-sm font-semibold text-text-primary sm:text-base"
              title={artifact.title || artifact.id}
            >
              {artifact.title || artifact.id}
            </h3>
            <div className="flex flex-wrap items-center gap-1.5">
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
            </div>
          </div>
          <div className="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-text-secondary">
            {artifact.targetUrl && (
              <span
                className="inline-flex min-w-0 max-w-full items-center gap-1.5"
                title={artifact.targetUrl}
              >
                <Globe size={14} className="shrink-0 text-brand-600" aria-hidden />
                <span className="truncate">{targetLabel(artifact.targetUrl)}</span>
              </span>
            )}
            <span title={Number.isNaN(addedAt.getTime()) ? undefined : addedAt.toLocaleString()}>
              {addedLabel}
            </span>
          </div>
        </div>
      </div>
      <div className="mt-4 flex flex-wrap items-center justify-between gap-3 border-t border-border pt-3">
        {revisionCount && revisionCount > 1 ? (
          <button
            type="button"
            onClick={onToggleRevisions}
            className="inline-flex min-h-8 items-center gap-1 rounded-full px-2 text-xs font-medium text-brand-600 hover:bg-brand-50 hover:text-brand-700 focus-visible:outline-2 focus-visible:outline-brand-500"
            aria-expanded={revisionsExpanded}
          >
            {revisionsExpanded ? <CaretDown size={13} /> : <CaretRight size={13} />}
            {revisionCount} revisions
          </button>
        ) : (
          <span />
        )}
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
      </div>
    </Card>
  );
}
