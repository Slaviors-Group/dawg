import { CircleNotch, Info, MagnifyingGlass } from "@phosphor-icons/react";
import { useEffect, useMemo, useState } from "react";
import type { DiagnosticEvidence, InspectResult } from "../lib/engine";
import { Badge } from "./ui/Badge";
import { Card, CardHeader, CardTitle } from "./ui/Card";
import { CodeBlock } from "./ui/CodeBlock";
import { EmptyState } from "./ui/EmptyState";
import { Input } from "./ui/Input";
import { Select } from "./ui/Select";

type WorkspaceTab = "timeline" | "console" | "network" | "errors" | "artifact";
type Evidence = Record<string, unknown>;

const WORKSPACE_TABS: Array<{ id: WorkspaceTab; label: string; keys: string[] }> = [
  { id: "timeline", label: "Timeline", keys: ["timeline", "events"] },
  { id: "console", label: "Console", keys: ["console", "consoleEntries", "logs"] },
  { id: "network", label: "Network", keys: ["network", "requests"] },
  { id: "errors", label: "Errors", keys: ["errors", "exceptions"] },
  { id: "artifact", label: "Artifact", keys: ["artifact", "artifacts", "evidence"] },
];

interface ReplayDiagnosticInspectorProps {
  manifest: InspectResult | null;
  diagnosticEvidence?: DiagnosticEvidence | null;
  loading?: boolean;
  error?: string | null;
  onExportHAR?: () => void;
  onRemoveDiagnostics?: () => void;
  onSeekReplay?: (offsetMs: number) => void;
  onCopyCurl?: (requestId: string) => void;
}

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const toEvidence = (value: unknown): Evidence[] => {
  if (Array.isArray(value)) return value.filter(isRecord);
  if (!isRecord(value)) return [];

  for (const key of ["samples", "entries", "events", "items", "data"]) {
    if (Array.isArray(value[key])) return value[key].filter(isRecord);
  }

  return Object.keys(value).length > 0 ? [value] : [];
};

const textValue = (value: unknown) => {
  if (typeof value === "string" || typeof value === "number" || typeof value === "boolean") {
    return String(value);
  }
  return "";
};

const firstText = (entry: Evidence, keys: string[]) => {
  for (const key of keys) {
    const value = textValue(entry[key]);
    if (value) return value;
  }
  return "";
};

const entryTitle = (entry: Evidence, index: number) => {
  const method = firstText(entry, ["method"]);
  const url = firstText(entry, ["url", "urlPath"]);
  if (method && url) return `${method} ${url}`;
  return (
    firstText(entry, ["message", "name", "title", "summary", "id", "type", "category"]) ||
    `Recorded evidence ${index + 1}`
  );
};

const entryDetail = (entry: Evidence) =>
  firstText(entry, ["description", "detail", "text", "url", "source", "filename"]);

const entryState = (entry: Evidence) =>
  firstText(entry, ["state", "status", "severity", "level", "outcome"]);

const entryTime = (entry: Evidence) =>
  firstText(entry, ["timestamp", "time", "occurredAt", "createdAt", "at"]);

const replayOffsetForEntry = (
  entry: Evidence,
  firstTimestamp?: number,
  durationMs?: number,
): number | null => {
  if (!Number.isFinite(firstTimestamp) || !Number.isFinite(durationMs)) return null;
  const timing = isRecord(entry.timing) ? entry.timing : null;
  const candidate = timing?.startedAt ?? entry.timestamp;
  const timestamp = typeof candidate === "number" ? candidate : Number(candidate);
  if (!Number.isFinite(timestamp)) return null;
  const offset = timestamp - (firstTimestamp as number);
  if (offset < 0 || offset > (durationMs as number)) return null;
  return Math.round(offset);
};

const formatReplayOffset = (offsetMs: number) => {
  const totalSeconds = Math.floor(offsetMs / 1000);
  return `${Math.floor(totalSeconds / 60)}:${String(totalSeconds % 60).padStart(2, "0")}.${String(offsetMs % 1000).padStart(3, "0")}`;
};

const badgeVariant = (state: string): "default" | "success" | "warning" | "error" | "info" => {
  const normalized = state.toLowerCase();
  if (/(error|fail|fatal|critical)/.test(normalized)) return "error";
  if (/(warn|pending|blocked)/.test(normalized)) return "warning";
  if (/(success|pass|complete|ok)/.test(normalized)) return "success";
  if (/(info|debug|trace|running)/.test(normalized)) return "info";
  return "default";
};

const tabEmptyCopy: Record<WorkspaceTab, { title: string; description: string }> = {
  timeline: {
    title: "No timeline evidence recorded",
    description: "This artifact does not provide timeline samples in its manifest diagnostics.",
  },
  console: {
    title: "No console evidence recorded",
    description: "Console entries are only shown when they are included in manifest diagnostics.",
  },
  network: {
    title: "No network evidence recorded",
    description: "No requests are inferred from the artifact or replay logs.",
  },
  errors: {
    title: "No error evidence recorded",
    description: "This manifest does not include captured error samples.",
  },
  artifact: {
    title: "No artifact evidence recorded",
    description: "This manifest does not include diagnostic artifact samples or layers.",
  },
};

export function ReplayDiagnosticInspector({
  manifest,
  diagnosticEvidence = null,
  loading = false,
  error = null,
  onExportHAR,
  onRemoveDiagnostics,
  onSeekReplay,
  onCopyCurl,
}: ReplayDiagnosticInspectorProps) {
  const [activeTab, setActiveTab] = useState<WorkspaceTab>("timeline");
  const [search, setSearch] = useState("");
  const [stateFilter, setStateFilter] = useState("all");
  const [selectedIndex, setSelectedIndex] = useState(0);

  const activeTabDefinition =
    WORKSPACE_TABS.find((tab) => tab.id === activeTab) ?? WORKSPACE_TABS[0];
  const diagnostics = isRecord(manifest?.diagnostics) ? manifest.diagnostics : null;
  const hasDiagnostics = diagnosticEvidence?.summary !== undefined || diagnostics !== null;

  const evidence = useMemo(() => {
    if (!manifest) return [];

    const evidenceEntries: Record<string, unknown> = {
      console: diagnosticEvidence?.console ?? [],
      network: diagnosticEvidence?.network ?? [],
      errors: diagnosticEvidence?.errors ?? [],
    };
    if (activeTab in evidenceEntries) {
      const entries = toEvidence(evidenceEntries[activeTab]);
      if (entries.length > 0) return entries;
    }
    for (const key of activeTabDefinition.keys) {
      const entries = toEvidence(diagnostics?.[key]);
      if (entries.length > 0) return entries;
    }

    // Layers are manifest-provided artifact evidence, not inferred diagnostics.
    if (activeTab === "artifact") return toEvidence(manifest.layers);
    return [];
  }, [activeTab, activeTabDefinition.keys, diagnostics, diagnosticEvidence, manifest]);

  const availableStates = useMemo(
    () =>
      Array.from(new Set(evidence.map(entryState).filter(Boolean))).sort((a, b) =>
        a.localeCompare(b),
      ),
    [evidence],
  );

  const filteredEvidence = useMemo(() => {
    const query = search.trim().toLowerCase();
    return evidence.filter((entry) => {
      if (stateFilter !== "all" && entryState(entry) !== stateFilter) return false;
      return !query || JSON.stringify(entry).toLowerCase().includes(query);
    });
  }, [evidence, search, stateFilter]);

  useEffect(() => {
    setSearch("");
    setStateFilter("all");
    setSelectedIndex(0);
  }, [evidence]);

  useEffect(() => {
    if (selectedIndex >= filteredEvidence.length) setSelectedIndex(0);
  }, [filteredEvidence.length, selectedIndex]);

  const selectedEvidence = filteredEvidence[selectedIndex];
  const selectedReplayOffset = selectedEvidence
    ? replayOffsetForEntry(
        selectedEvidence,
        diagnosticEvidence?.timeline?.firstTimestamp,
        diagnosticEvidence?.timeline?.durationMs,
      )
    : null;
  const emptyCopy = tabEmptyCopy[activeTab];

  return (
    <Card noPad className="overflow-hidden">
      <CardHeader className="px-5 pt-5">
        <CardTitle
          title="Replay Diagnostics"
          subtitle="Evidence recorded in the selected artifact manifest"
        />
        <div className="flex items-center gap-2">
          {onExportHAR && diagnosticEvidence?.network.length ? (
            <button type="button" onClick={onExportHAR} className="text-xs font-medium text-brand-600 hover:text-brand-700">
              Export HAR
            </button>
          ) : null}
          {onRemoveDiagnostics && diagnosticEvidence?.summary ? (
            <button type="button" onClick={onRemoveDiagnostics} className="text-xs font-medium text-brand-600 hover:text-brand-700">
              Remove evidence…
            </button>
          ) : null}
          {manifest && (
            <Badge variant={hasDiagnostics ? "info" : "default"} size="sm" dot={hasDiagnostics}>
              {hasDiagnostics ? "Manifest diagnostics" : "Legacy manifest"}
            </Badge>
          )}
        </div>
      </CardHeader>

      <div className="border-b border-border px-3 sm:px-5" role="tablist" aria-label="Replay diagnostics">
        <div className="flex gap-1 overflow-x-auto">
          {WORKSPACE_TABS.map((tab) => {
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                type="button"
                role="tab"
                aria-selected={isActive}
                onClick={() => setActiveTab(tab.id)}
                className={[
                  "shrink-0 border-b-2 px-3 py-2.5 text-xs font-medium transition-colors",
                  isActive
                    ? "border-brand-500 text-brand-600"
                    : "border-transparent text-text-tertiary hover:text-text-primary",
                ].join(" ")}
              >
                {tab.label}
              </button>
            );
          })}
        </div>
      </div>

      {loading ? (
        <EmptyState
          icon={<CircleNotch size={30} className="animate-spin" />}
          title="Loading artifact diagnostics"
          description="Reading the selected artifact manifest…"
        />
      ) : error ? (
        <EmptyState
          icon={<Info size={32} weight="light" />}
          title="Diagnostics unavailable"
          description={error}
        />
      ) : !manifest ? (
        <EmptyState
          icon={<Info size={32} weight="light" />}
          title="Select an artifact to inspect diagnostics"
          description="The workspace displays only evidence available in that artifact's manifest."
        />
      ) : (
        <div className="p-5">
          {!hasDiagnostics && activeTab !== "artifact" && (
            <div className="mb-4 flex gap-2 rounded-md border border-info-border bg-info-bg px-3 py-2 text-xs text-info-text">
              <Info size={15} className="mt-0.5 shrink-0" />
              <p>
                This legacy manifest does not contain diagnostics. DAWG will not infer timeline,
                console, network, or error data.
              </p>
            </div>
          )}

          {evidence.length === 0 ? (
            <EmptyState
              icon={<Info size={32} weight="light" />}
              title={emptyCopy.title}
              description={emptyCopy.description}
            />
          ) : (
            <div className="flex flex-col gap-4">
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-[minmax(0,1fr)_12rem]">
                <Input
                  id="diagnostic-search"
                  label="Search evidence"
                  placeholder="Search recorded fields..."
                  value={search}
                  onChange={(event) => setSearch(event.target.value)}
                  iconLeft={<MagnifyingGlass size={14} />}
                />
                <Select
                  id="diagnostic-state-filter"
                  label="State"
                  value={stateFilter}
                  onChange={setStateFilter}
                  options={[
                    { value: "all", label: `All states (${evidence.length})` },
                    ...availableStates.map((state) => ({ value: state, label: state })),
                  ]}
                />
              </div>

              {filteredEvidence.length === 0 ? (
                <EmptyState
                  icon={<MagnifyingGlass size={30} weight="light" />}
                  title="No evidence matches these filters"
                  description="Try clearing the search or choosing another recorded state."
                />
              ) : (
                <div className="grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(16rem,0.85fr)]">
                  <div className="max-h-96 overflow-y-auto rounded-md border border-border divide-y divide-border">
                    {filteredEvidence.map((entry, index) => {
                      const state = entryState(entry);
                      const time = entryTime(entry);
                      const isSelected = index === selectedIndex;
                      return (
                        <button
                          key={`${entryTitle(entry, index)}-${index}`}
                          type="button"
                          onClick={() => setSelectedIndex(index)}
                          className={[
                            "flex w-full items-start gap-3 px-3 py-3 text-left transition-colors",
                            isSelected ? "bg-brand-500/10" : "hover:bg-surface-hover",
                          ].join(" ")}
                        >
                          <div className="min-w-0 flex-1">
                            <p className="truncate text-sm font-medium text-text-primary">
                              {entryTitle(entry, index)}
                            </p>
                            {(entryDetail(entry) || time) && (
                              <p className="mt-0.5 truncate text-xs text-text-tertiary">
                                {[entryDetail(entry), time].filter(Boolean).join(" · ")}
                              </p>
                            )}
                          </div>
                          {state && (
                            <Badge variant={badgeVariant(state)} size="sm" className="shrink-0">
                              {state}
                            </Badge>
                          )}
                        </button>
                      );
                    })}
                  </div>

                  {selectedEvidence && (
                    <div className="min-w-0 rounded-md border border-border bg-canvas-subtle p-3">
                      <div className="mb-3 flex items-start justify-between gap-3">
                        <div className="min-w-0">
                          <p className="text-[10px] font-medium uppercase tracking-wider text-text-tertiary">
                            Evidence details
                          </p>
                          <p className="mt-1 wrap-break-word text-sm font-medium text-text-primary">
                            {entryTitle(selectedEvidence, selectedIndex)}
                          </p>
                        </div>
                        {entryState(selectedEvidence) && (
                          <Badge variant={badgeVariant(entryState(selectedEvidence))} size="sm">
                            {entryState(selectedEvidence)}
                          </Badge>
                        )}
                      </div>
                      <div className="mb-3 flex flex-wrap gap-x-3 gap-y-2">
                        {activeTab === "network" && onCopyCurl && typeof selectedEvidence.requestId === "string" ? (
                          <button type="button" onClick={() => onCopyCurl(selectedEvidence.requestId as string)} className="text-xs font-medium text-brand-600 hover:text-brand-700">
                            Copy as cURL
                          </button>
                        ) : null}
                        {onSeekReplay && selectedReplayOffset !== null ? (
                          <button type="button" onClick={() => onSeekReplay(selectedReplayOffset)} className="text-xs font-medium text-brand-600 hover:text-brand-700">
                            Seek replay to approximately {formatReplayOffset(selectedReplayOffset)}
                          </button>
                        ) : null}
                      </div>
                      {diagnosticEvidence?.timeline && selectedReplayOffset === null ? (
                        <p className="mb-3 text-xs text-text-tertiary">No replay correlation is available for this record.</p>
                      ) : null}
                      <CodeBlock
                        code={JSON.stringify(selectedEvidence, null, 2)}
                        language="json"
                        maxHeight="16rem"
                      />
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </Card>
  );
}
