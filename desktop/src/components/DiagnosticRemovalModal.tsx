import { useEffect, useMemo, useState } from "react";
import type { DiagnosticCategory, DiagnosticEvidence } from "../lib/engine";
import { Button } from "./ui/Button";
import { Input } from "./ui/Input";
import { Modal } from "./ui/Modal";

interface DiagnosticRemovalModalProps {
  open: boolean;
  evidence: DiagnosticEvidence | null;
  defaultDirectoryName: string;
  onClose: () => void;
  onChooseParentDirectory: () => Promise<string | null>;
  onConfirm: (request: {
    categories: DiagnosticCategory[];
    bodyRefs: string[];
    parentDirectory: string;
    directoryName: string;
  }) => Promise<void>;
}

const categories: Array<{ id: DiagnosticCategory; label: string; description: string }> = [
  { id: "console", label: "Console", description: "Remove retained console records." },
  {
    id: "network",
    label: "Network",
    description: "Remove network records and their retained bodies.",
  },
  { id: "errors", label: "Errors", description: "Remove retained error records." },
  { id: "device", label: "Device", description: "Remove the captured browser device profile." },
  {
    id: "bodies",
    label: "All bodies",
    description: "Remove all retained request and response bodies.",
  },
];

export function DiagnosticRemovalModal({
  open,
  evidence,
  defaultDirectoryName,
  onClose,
  onChooseParentDirectory,
  onConfirm,
}: DiagnosticRemovalModalProps) {
  const [selectedCategories, setSelectedCategories] = useState<DiagnosticCategory[]>([]);
  const [selectedBodyRefs, setSelectedBodyRefs] = useState<string[]>([]);
  const [parentDirectory, setParentDirectory] = useState("");
  const [directoryName, setDirectoryName] = useState(defaultDirectoryName);
  const [confirmed, setConfirmed] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const bodyRefs = useMemo(
    () => Object.keys(evidence?.bodies ?? {}).sort((a, b) => a.localeCompare(b)),
    [evidence],
  );
  const removesAllBodies =
    selectedCategories.includes("network") || selectedCategories.includes("bodies");

  useEffect(() => {
    if (!open) return;
    setSelectedCategories([]);
    setSelectedBodyRefs([]);
    setParentDirectory("");
    setDirectoryName(defaultDirectoryName);
    setConfirmed(false);
  }, [defaultDirectoryName, open]);

  const toggleCategory = (category: DiagnosticCategory) => {
    setSelectedCategories((current) =>
      current.includes(category)
        ? current.filter((value) => value !== category)
        : [...current, category],
    );
  };

  const chooseParent = async () => {
    const selected = await onChooseParentDirectory();
    if (selected) setParentDirectory(selected);
  };

  const validDirectoryName = directoryName.trim() !== "" && !/[\\/]/.test(directoryName);
  const hasRemoval = selectedCategories.length > 0 || selectedBodyRefs.length > 0;
  const canSubmit = confirmed && hasRemoval && parentDirectory !== "" && validDirectoryName;

  const close = () => {
    if (!submitting) onClose();
  };

  const submit = async () => {
    if (!canSubmit) return;
    setSubmitting(true);
    try {
      await onConfirm({
        categories: selectedCategories,
        bodyRefs: removesAllBodies ? [] : selectedBodyRefs,
        parentDirectory,
        directoryName: directoryName.trim(),
      });
      onClose();
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      open={open}
      onClose={close}
      title="Create reviewed diagnostic copy"
      subtitle="The original artifact stays unchanged. Removed evidence cannot be restored in the new copy."
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="secondary" size="sm" onClick={close} disabled={submitting}>
            Cancel
          </Button>
          <Button
            variant="danger"
            size="sm"
            onClick={() => void submit()}
            disabled={!canSubmit}
            loading={submitting}
          >
            Create reviewed copy
          </Button>
        </div>
      }
    >
      <div className="flex flex-col gap-4 text-sm">
        <fieldset className="flex flex-col gap-2">
          <legend className="text-xs font-medium text-text-primary">
            Remove complete evidence categories
          </legend>
          {categories.map((category) => (
            <label
              key={category.id}
              className="flex cursor-pointer items-start gap-2 rounded-md border border-border px-3 py-2 text-xs"
            >
              <input
                type="checkbox"
                checked={selectedCategories.includes(category.id)}
                onChange={() => toggleCategory(category.id)}
                className="mt-0.5 accent-[--color-brand-500]"
              />
              <span>
                <span className="font-medium text-text-primary">{category.label}</span>
                <span className="block text-text-tertiary">{category.description}</span>
              </span>
            </label>
          ))}
        </fieldset>

        {bodyRefs.length > 0 && !removesAllBodies ? (
          <fieldset className="flex flex-col gap-2">
            <legend className="text-xs font-medium text-text-primary">
              Or remove selected retained bodies
            </legend>
            <div className="max-h-36 overflow-y-auto rounded-md border border-border divide-y divide-border">
              {bodyRefs.map((ref) => (
                <label
                  key={ref}
                  className="flex cursor-pointer items-start gap-2 px-3 py-2 text-xs"
                >
                  <input
                    type="checkbox"
                    checked={selectedBodyRefs.includes(ref)}
                    onChange={() =>
                      setSelectedBodyRefs((current) =>
                        current.includes(ref)
                          ? current.filter((value) => value !== ref)
                          : [...current, ref],
                      )
                    }
                    className="mt-0.5 accent-[--color-brand-500]"
                  />
                  <span className="break-all text-text-secondary">{ref}</span>
                </label>
              ))}
            </div>
          </fieldset>
        ) : null}

        <div className="grid grid-cols-1 gap-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-end">
          <Input
            id="reviewed-artifact-name"
            label="New artifact directory name"
            value={directoryName}
            onChange={(event) => setDirectoryName(event.target.value)}
            error={
              !validDirectoryName ? "Use a non-empty name without path separators." : undefined
            }
          />
          <Button
            variant="secondary"
            size="sm"
            onClick={() => void chooseParent()}
            disabled={submitting}
          >
            Choose parent folder
          </Button>
        </div>
        <p className="break-all rounded-md border border-border bg-canvas-subtle p-2 text-xs text-text-secondary">
          {parentDirectory
            ? `New copy: ${parentDirectory} / ${directoryName || "…"}`
            : "Choose an existing parent folder for the new artifact directory."}
        </p>
        <label className="flex cursor-pointer items-start gap-2 text-xs text-text-primary">
          <input
            type="checkbox"
            checked={confirmed}
            onChange={(event) => setConfirmed(event.target.checked)}
            className="mt-0.5 accent-[--color-brand-500]"
          />
          <span>
            I reviewed the selected evidence removal and understand this creates a new artifact with
            new digests and no inherited provenance.
          </span>
        </label>
      </div>
    </Modal>
  );
}
