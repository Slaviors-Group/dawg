import { X } from "@phosphor-icons/react";
import { AnimatePresence, motion } from "framer-motion";

import { type ReactNode, useEffect } from "react";
import { createPortal } from "react-dom";
import { Button } from "./Button";

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  subtitle?: string;
  children: ReactNode;

  footerLeft?: ReactNode;

  footer?: ReactNode;
  maxWidth?: "sm" | "md" | "lg";
}

const maxWidthClass = {
  sm: "max-w-sm",
  md: "max-w-lg",
  lg: "max-w-2xl",
};

export function Modal({
  open,
  onClose,
  title,
  subtitle,
  children,
  footerLeft,
  footer,
  maxWidth = "md",
}: ModalProps) {
  // Escape key dismissal
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [open, onClose]);

  return createPortal(
    <AnimatePresence>
      {open && (
        <motion.div
          key="modal-backdrop"
          className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.15 }}
        >
          {/* Backdrop */}
          <button
            type="button"
            aria-label="Close modal"
            className="absolute inset-0 bg-[hsl(240_22%_15%_/_0.45)] backdrop-blur-[5px]"
            onClick={onClose}
          />

          {/* Dialog */}
          <motion.dialog
            open
            aria-modal
            aria-labelledby="modal-title"
            className={[
              "relative m-0 w-full bg-surface rounded-3xl",
              "border border-border shadow-modal",
              "flex min-h-0 flex-col max-h-[calc(100dvh-1rem)] overflow-hidden sm:max-h-[calc(100dvh-3rem)]",
              maxWidthClass[maxWidth],
            ].join(" ")}
            initial={{ opacity: 0, scale: 0.95, y: 12 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: 12 }}
            transition={{ type: "spring", duration: 0.3, bounce: 0.2 }}
          >
            {/* Header */}
            <div className="flex shrink-0 items-start justify-between gap-3 px-5 pt-5 pb-3 sm:gap-4 sm:px-7 sm:pt-7 sm:pb-4">
              <div className="min-w-0">
                <h3
                  id="modal-title"
                  className="text-xl font-semibold tracking-tight text-text-primary"
                >
                  {title}
                </h3>
                {subtitle && (
                  <p className="mt-1 text-sm text-text-tertiary wrap-anywhere">{subtitle}</p>
                )}
              </div>
              <button
                type="button"
                onClick={onClose}
                aria-label="Close dialog"
                className="grid size-9 shrink-0 place-items-center rounded-full bg-canvas-subtle text-text-tertiary transition-colors hover:bg-brand-100 hover:text-brand-700 focus-visible:outline-2 focus-visible:outline-brand-500"
              >
                <X size={17} />
              </button>
            </div>

            {/* Body */}
            <div className="min-h-0 min-w-0 flex-1 overflow-y-auto px-5 py-3 sm:px-7 sm:py-4">
              {children}
            </div>

            {/* Footer */}
            {footer !== undefined ? (
              <div className="shrink-0 border-t border-border/70 bg-canvas-subtle/60 px-5 py-4 sm:px-7">
                {footer}
              </div>
            ) : (
              <div className="flex shrink-0 items-center justify-between gap-3 border-t border-border/70 bg-canvas-subtle/60 px-5 py-4 sm:px-7">
                <div>{footerLeft}</div>
                <Button variant="secondary" size="sm" onClick={onClose}>
                  Close
                </Button>
              </div>
            )}
          </motion.dialog>
        </motion.div>
      )}
    </AnimatePresence>,
    document.body,
  );
}
