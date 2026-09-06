/**
 * Toast notification system.
 * Uses a ToastProvider + useToast() hook.
 * Toasts stack at bottom-right with Framer Motion entry/exit.
 */
import {
  type ReactNode,
  createContext,
  useCallback,
  useContext,
  useRef,
  useState,
} from "react";
import {
  CheckCircle,
  Info,
  Warning,
  XCircle,
  X,
} from "@phosphor-icons/react";
import { AnimatePresence, motion } from "framer-motion";

export type ToastVariant = "success" | "error" | "warning" | "info";

interface Toast {
  id: string;
  variant: ToastVariant;
  title: string;
  body?: string;
}

interface ToastContextType {
  toast: (opts: Omit<Toast, "id">) => void;
  success: (title: string, body?: string) => void;
  error: (title: string, body?: string) => void;
  warning: (title: string, body?: string) => void;
  info: (title: string, body?: string) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

const variantStyle: Record<ToastVariant, { icon: ReactNode; classes: string }> =
  {
    success: {
      icon: <CheckCircle size={16} weight="fill" />,
      classes:
        "bg-success-bg border-success-border text-success-text",
    },
    error: {
      icon: <XCircle size={16} weight="fill" />,
      classes:
        "bg-error-bg border-error-border text-error-text",
    },
    warning: {
      icon: <Warning size={16} weight="fill" />,
      classes:
        "bg-warning-bg border-warning-border text-warning-text",
    },
    info: {
      icon: <Info size={16} weight="fill" />,
      classes:
        "bg-info-bg border-info-border text-info-text",
    },
  };

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const timers = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map());

  const dismiss = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
    const timer = timers.current.get(id);
    if (timer) clearTimeout(timer);
    timers.current.delete(id);
  }, []);

  const toast = useCallback(
    (opts: Omit<Toast, "id">) => {
      const id = crypto.randomUUID();
      setToasts((prev) => [...prev.slice(-4), { ...opts, id }]);
      const timer = setTimeout(() => dismiss(id), 4500);
      timers.current.set(id, timer);
    },
    [dismiss]
  );

  const helpers: ToastContextType = {
    toast,
    success: (title, body) => toast({ variant: "success", title, body }),
    error:   (title, body) => toast({ variant: "error",   title, body }),
    warning: (title, body) => toast({ variant: "warning", title, body }),
    info:    (title, body) => toast({ variant: "info",    title, body }),
  };

  return (
    <ToastContext.Provider value={helpers}>
      {children}
      {/* Toast portal */}
      <div
        aria-live="polite"
        aria-label="Notifications"
        className="fixed bottom-4 right-4 z-[70] flex flex-col gap-2 w-80 max-w-[calc(100vw-2rem)]"
      >
        <AnimatePresence mode="popLayout">
          {toasts.map((t) => {
            const { icon, classes } = variantStyle[t.variant];
            return (
              <motion.div
                key={t.id}
                layout
                initial={{ opacity: 0, y: 16, scale: 0.97 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, x: 24, scale: 0.95 }}
                transition={{ type: "spring", duration: 0.35, bounce: 0.25 }}
                className={[
                  "flex items-start gap-2.5 p-3.5 rounded-lg",
                  "border shadow-modal",
                  classes,
                ].join(" ")}
              >
                <span className="shrink-0 mt-0.5">{icon}</span>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-semibold leading-tight">
                    {t.title}
                  </p>
                  {t.body && (
                    <p className="text-xs mt-0.5 opacity-80 leading-relaxed">
                      {t.body}
                    </p>
                  )}
                </div>
                <button
                  type="button"
                  aria-label="Dismiss"
                  onClick={() => dismiss(t.id)}
                  className="shrink-0 p-0.5 opacity-60 hover:opacity-100 rounded transition-opacity"
                >
                  <X size={12} />
                </button>
              </motion.div>
            );
          })}
        </AnimatePresence>
      </div>
    </ToastContext.Provider>
  );
}

export function useToast() {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error("useToast must be used within ToastProvider");
  return ctx;
}
