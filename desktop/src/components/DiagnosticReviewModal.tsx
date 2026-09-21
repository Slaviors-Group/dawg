import { useState } from "react";
import type { DiagnosticEvidence } from "../lib/engine";
import { Button } from "./ui/Button";
import { Modal } from "./ui/Modal";

interface DiagnosticReviewModalProps {
  open: boolean;
  action: "artifact" | "har" | "curl" | null;
  evidence: DiagnosticEvidence | null;
  onClose: () => void;
  onConfirm: () => Promise<void>;
}

const actionLabels = {
  artifact: "Export artifact",
  har: "Export sanitized HAR",
  curl: "Copy reviewed cURL",
};

export function DiagnosticReviewModal({
  open,
  action,
  evidence,
  onClose,
  onConfirm,
}: DiagnosticReviewModalProps) {
  const [confirmed, setConfirmed] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const summary = evidence?.summary ?? {};

  const close = () => {
    if (submitting) return;
    setConfirmed(false);
    onClose();
  };

  const submit = async () => {
    if (!confirmed || !action) return;
    setSubmitting(true);
    try {
      await onConfirm();
      setConfirmed(false);
      onClose();
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      open={open}
      onClose={close}
      title="Review diagnostic evidence"
      subtitle="Confirm the sanitized evidence before it leaves this device."
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="secondary" size="sm" onClick={close} disabled={submitting}>
            Cancel
          </Button>
          <Button
            variant="primary"
            size="sm"
            onClick={() => void submit()}
            disabled={!confirmed}
            loading={submitting}
          >
            {action ? actionLabels[action] : "Continue"}
          </Button>
        </div>
      }
    >
      <div className="flex flex-col gap-4 text-sm">
        <p className="text-text-secondary">
          DAWG exports only evidence retained in the artifact after sanitization. Empty, blocked,
          unavailable, and truncated values are not reconstructed.
        </p>
        <dl className="grid grid-cols-2 gap-x-4 gap-y-2 rounded-md border border-border bg-canvas-subtle p-3 text-xs">
          <div>
            <dt className="text-text-tertiary">Profile</dt>
            <dd className="font-medium text-text-primary">{String(summary.profile ?? "legacy")}</dd>
          </div>
          <div>
            <dt className="text-text-tertiary">Console records</dt>
            <dd className="font-medium text-text-primary">
              {String(summary.consoleEvents ?? evidence?.console.length ?? 0)}
            </dd>
          </div>
          <div>
            <dt className="text-text-tertiary">Network records</dt>
            <dd className="font-medium text-text-primary">
              {String(summary.networkEvents ?? evidence?.network.length ?? 0)}
            </dd>
          </div>
          <div>
            <dt className="text-text-tertiary">Error records</dt>
            <dd className="font-medium text-text-primary">
              {String(summary.errorEvents ?? evidence?.errors.length ?? 0)}
            </dd>
          </div>
          <div>
            <dt className="text-text-tertiary">Blocked bodies</dt>
            <dd className="font-medium text-text-primary">{String(summary.blockedBodies ?? 0)}</dd>
          </div>
          <div>
            <dt className="text-text-tertiary">Truncated bodies</dt>
            <dd className="font-medium text-text-primary">
              {String(summary.truncatedBodies ?? 0)}
            </dd>
          </div>
        </dl>
        <label className="flex cursor-pointer items-start gap-2 text-xs text-text-primary">
          <input
            type="checkbox"
            checked={confirmed}
            onChange={(event) => setConfirmed(event.target.checked)}
            className="mt-0.5 accent-[--color-brand-500]"
          />
          <span>
            I reviewed the retained diagnostic evidence and understand that a copied cURL may mutate
            a live service.
          </span>
        </label>
      </div>
    </Modal>
  );
}
