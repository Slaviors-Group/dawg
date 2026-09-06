/**
 * Modal — full-screen backdrop + centered dialog.
 * Manages focus trap and Escape key dismissal.
 */
import { type ReactNode, useEffect } from "react";
import { X } from "@phosphor-icons/react";
import { AnimatePresence, motion } from "framer-motion";
import { Button } from "./Button";

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  subtitle?: string;
  children: ReactNode;
  /** Extra content to render in the footer, to the left of the Close button */
  footerLeft?: ReactNode;
  /** Replace the default Close footer with a custom footer */
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

  return (
    <AnimatePresence>
      {open && (
        <motion.div
          key="modal-backdrop"
          className="fixed inset-0 z-[50] flex items-center justify-center p-4 sm:p-6"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.15 }}
        >
          {/* Backdrop */}
          <div
            className="absolute inset-0 bg-text-primary/40 backdrop-blur-sm"
            onClick={onClose}
            aria-hidden
          />

          {/* Dialog */}
          <motion.div
            role="dialog"
            aria-modal
            aria-labelledby="modal-title"
            className={[
              "relative w-full bg-surface rounded-xl",
              "border border-border shadow-modal",
              "flex flex-col max-h-[88vh] overflow-hidden",
              maxWidthClass[maxWidth],
            ].join(" ")}
            initial={{ opacity: 0, scale: 0.95, y: 12 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: 12 }}
            transition={{ type: "spring", duration: 0.3, bounce: 0.2 }}
          >
            {/* Header */}
            <div className="flex items-start justify-between gap-4 px-6 py-5 border-b border-border shrink-0">
              <div>
                <h3
                  id="modal-title"
                  className="text-base font-semibold text-text-primary"
                >
                  {title}
                </h3>
                {subtitle && (
                  <p className="text-xs text-text-tertiary mt-0.5">
                    {subtitle}
                  </p>
                )}
              </div>
              <button
                type="button"
                onClick={onClose}
                aria-label="Close dialog"
                className="p-1 text-text-tertiary hover:text-text-primary rounded-sm transition-colors duration-[--duration-fast] shrink-0"
              >
                <X size={16} />
              </button>
            </div>

            {/* Body */}
            <div className="flex-1 overflow-y-auto px-6 py-5">{children}</div>

            {/* Footer */}
            {footer !== undefined ? (
              <div className="px-6 py-4 border-t border-border shrink-0">
                {footer}
              </div>
            ) : (
              <div className="flex items-center justify-between gap-3 px-6 py-4 border-t border-border shrink-0">
                <div>{footerLeft}</div>
                <Button variant="secondary" size="sm" onClick={onClose}>
                  Close
                </Button>
              </div>
            )}
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
