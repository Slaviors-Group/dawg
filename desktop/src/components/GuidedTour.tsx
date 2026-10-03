import { ArrowLeft, ArrowRight, X } from "@phosphor-icons/react";
import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

import { Button } from "./ui/Button";

export type GuideTab = "dashboard" | "capture" | "artifacts" | "replay" | "editor";

const STEPS: { tab: GuideTab; target: string; title: string; description: string }[] = [
  {
    tab: "dashboard",
    target: "dashboard-start",
    title: "Start from the Dashboard",
    description:
      "Start capture opens the recording setup. You can also browse saved artifacts from here.",
  },
  {
    tab: "dashboard",
    target: "dashboard-status",
    title: "Workspace status",
    description:
      "Check whether the engine is ready, see where artifacts came from, and review Doctor compatibility.",
  },
  {
    tab: "dashboard",
    target: "dashboard-summary",
    title: "Your activity at a glance",
    description: "These cards count local artifacts, review flags, and recorded diagnostics.",
  },
  {
    tab: "dashboard",
    target: "dashboard-latest",
    title: "Latest artifact",
    description: "Open your newest artifact in Replay or Editor. View all opens the full catalog.",
  },
  {
    tab: "capture",
    target: "capture-target",
    title: "Choose a target",
    description:
      "Enter the web app URL. An artifact name is optional. Install the DAWG browser extension before recording.",
  },
  {
    tab: "capture",
    target: "capture-profile-safe",
    title: "Safe capture",
    description:
      "Safe is the default. It sanitizes replay data and records limited console, error, and network metadata without requesting response bodies.",
  },
  {
    tab: "capture",
    target: "capture-profile-enhanced",
    title: "Enhanced Diagnostics",
    description:
      "This adds browser debugger evidence and eligible text response bodies. It may retain more data, so review the warning and confirm before recording.",
  },
  {
    tab: "capture",
    target: "capture-action",
    title: "Record the issue",
    description:
      "Select Start Capture, reproduce the issue in the browser, then select Stop capture. DAWG packages the session as an artifact.",
  },
  {
    tab: "artifacts",
    target: "artifacts-search",
    title: "Find an artifact",
    description:
      "Search by name, URL, ID, or path. Filters narrow the catalog. Press Enter to inspect the first match.",
  },
  {
    tab: "artifacts",
    target: "artifacts-import",
    title: "Import a .dawg file",
    description: "Select Import to choose a .dawg archive, or drop one onto this page.",
  },
  {
    tab: "artifacts",
    target: "artifacts-list",
    title: "Inspect and export",
    description:
      "Each artifact row offers Inspect, Replay, Edit, Export, and Delete. Export saves a shareable .dawg file. Related revisions stay grouped.",
  },
  {
    tab: "replay",
    target: "replay-choose",
    title: "Choose a replay",
    description:
      "Select an artifact, then Run Replay. DAWG opens the recorded session in Chromium. Import is available here too.",
  },
  {
    tab: "replay",
    target: "replay-controls",
    title: "Control playback",
    description:
      "Play or pause, seek on the timeline, and change speed. Review captured evidence and jump between flags when available.",
  },
  {
    tab: "editor",
    target: "editor-selection",
    title: "Review in Editor",
    description:
      "Select an artifact and choose Edit. Add or change point and range flags. Save draft keeps your work; Save artifact publishes a new revision and preserves the source.",
  },
];

type Placement = "top" | "bottom" | "left" | "right";

const clamp = (value: number, min: number, max: number) => Math.min(Math.max(value, min), max);

interface GuidedTourProps {
  step: number;
  activeTab: GuideTab | "diff" | "policy";
  onStepChange: (step: number) => void;
  onNavigate: (tab: GuideTab) => void;
  onFinish: () => void;
}

export function GuidedTour({
  step,
  activeTab,
  onStepChange,
  onNavigate,
  onFinish,
}: GuidedTourProps) {
  const activeStep = STEPS[step] ?? STEPS[0];
  const targetKey = `${step}:${activeStep.target}`;
  const [targetState, setTargetState] = useState<{ key: string; rect: DOMRect | null } | null>(
    null,
  );
  const [popoverHeight, setPopoverHeight] = useState(260);
  const [viewport, setViewport] = useState(() => ({
    width: window.innerWidth,
    height: window.innerHeight,
  }));
  const popoverRef = useRef<HTMLDialogElement>(null);
  const nextRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    const previousFocus =
      document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        onFinish();
        return;
      }
      if (
        event.key === "Enter" &&
        !event.repeat &&
        !event.isComposing &&
        !event.altKey &&
        !event.ctrlKey &&
        !event.metaKey &&
        !event.shiftKey
      ) {
        const target = event.target;
        if (
          !(target instanceof Element) ||
          !popoverRef.current?.contains(target) ||
          !target.closest("button, a, input, select, textarea, [contenteditable='true']")
        ) {
          event.preventDefault();
          nextRef.current?.click();
        }
        return;
      }
      if (event.key !== "Tab" || !popoverRef.current) return;
      const buttons = Array.from(
        popoverRef.current.querySelectorAll<HTMLButtonElement>("button:not(:disabled)"),
      );
      const first = buttons[0];
      const last = buttons[buttons.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last?.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first?.focus();
      }
    };
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("keydown", onKeyDown);
      previousFocus?.focus();
    };
  }, [onFinish]);

  useEffect(() => {
    const updateViewport = () =>
      setViewport({ width: window.innerWidth, height: window.innerHeight });
    window.addEventListener("resize", updateViewport);
    return () => window.removeEventListener("resize", updateViewport);
  }, []);

  useEffect(() => {
    if (activeTab !== activeStep.tab) {
      onNavigate(activeStep.tab);
      return;
    }

    let retryTimer: ReturnType<typeof setTimeout> | undefined;
    let animationFrame: number | undefined;
    let target: HTMLElement | null = null;
    let observer: ResizeObserver | undefined;
    const updateRect = () => {
      if (target) setTargetState({ key: targetKey, rect: target.getBoundingClientRect() });
    };
    const findTarget = (attempt: number) => {
      target = document.querySelector<HTMLElement>(`[data-tour="${activeStep.target}"]`);
      if (!target) {
        if (attempt < 30) retryTimer = setTimeout(() => findTarget(attempt + 1), 50);
        else setTargetState({ key: targetKey, rect: null });
        return;
      }
      target.scrollIntoView({ block: "center", inline: "nearest" });
      const startedAt = performance.now();
      let previousRect: DOMRect | null = null;
      let stableFrames = 0;
      const measureUntilSettled = () => {
        if (!target?.isConnected) {
          setTargetState({ key: targetKey, rect: null });
          return;
        }
        const rect = target.getBoundingClientRect();
        const unchanged =
          previousRect !== null &&
          Math.abs(rect.left - previousRect.left) < 0.5 &&
          Math.abs(rect.top - previousRect.top) < 0.5 &&
          Math.abs(rect.width - previousRect.width) < 0.5 &&
          Math.abs(rect.height - previousRect.height) < 0.5;
        stableFrames = unchanged ? stableFrames + 1 : 0;
        previousRect = rect;
        if (
          (performance.now() - startedAt >= 260 && stableFrames >= 2) ||
          performance.now() - startedAt >= 1200
        ) {
          setTargetState({ key: targetKey, rect });
          window.addEventListener("scroll", updateRect, true);
          window.addEventListener("resize", updateRect);
          if (typeof ResizeObserver !== "undefined") {
            observer = new ResizeObserver(updateRect);
            observer.observe(target);
          }
          return;
        }
        animationFrame = requestAnimationFrame(measureUntilSettled);
      };
      animationFrame = requestAnimationFrame(measureUntilSettled);
    };
    findTarget(0);
    return () => {
      if (retryTimer) clearTimeout(retryTimer);
      if (animationFrame !== undefined) cancelAnimationFrame(animationFrame);
      window.removeEventListener("scroll", updateRect, true);
      window.removeEventListener("resize", updateRect);
      observer?.disconnect();
    };
  }, [activeStep, activeTab, onNavigate, targetKey]);

  useEffect(() => {
    if (!popoverRef.current || typeof ResizeObserver === "undefined") return;
    const observer = new ResizeObserver(() => {
      if (popoverRef.current) setPopoverHeight(popoverRef.current.getBoundingClientRect().height);
    });
    observer.observe(popoverRef.current);
    return () => observer.disconnect();
  }, []);

  const targetReady = activeTab === activeStep.tab && targetState?.key === targetKey;

  useEffect(() => {
    if (targetReady) nextRef.current?.focus();
  }, [targetReady]);

  const targetRect = targetReady ? targetState.rect : null;
  const spotlight = targetRect
    ? {
        left: clamp(targetRect.left - 6, 0, viewport.width),
        top: clamp(targetRect.top - 6, 0, viewport.height),
        right: clamp(targetRect.right + 6, 0, viewport.width),
        bottom: clamp(targetRect.bottom + 6, 0, viewport.height),
      }
    : null;
  const popoverWidth = Math.min(352, viewport.width - 24);
  const gap = 16;
  let placement: Placement = "bottom";
  let popoverLeft = (viewport.width - popoverWidth) / 2;
  let popoverTop = (viewport.height - popoverHeight) / 2;

  if (spotlight) {
    if (viewport.height - spotlight.bottom >= popoverHeight + gap + 12) placement = "bottom";
    else if (spotlight.top >= popoverHeight + gap + 12) placement = "top";
    else if (viewport.width - spotlight.right >= popoverWidth + gap + 12) placement = "right";
    else if (spotlight.left >= popoverWidth + gap + 12) placement = "left";
    else placement = viewport.height - spotlight.bottom >= spotlight.top ? "bottom" : "top";

    if (placement === "bottom" || placement === "top") {
      popoverLeft = clamp(
        (spotlight.left + spotlight.right - popoverWidth) / 2,
        12,
        viewport.width - popoverWidth - 12,
      );
      popoverTop = clamp(
        placement === "bottom" ? spotlight.bottom + gap : spotlight.top - popoverHeight - gap,
        12,
        viewport.height - popoverHeight - 12,
      );
    } else {
      popoverLeft = clamp(
        placement === "right" ? spotlight.right + gap : spotlight.left - popoverWidth - gap,
        12,
        viewport.width - popoverWidth - 12,
      );
      popoverTop = clamp(
        (spotlight.top + spotlight.bottom - popoverHeight) / 2,
        12,
        viewport.height - popoverHeight - 12,
      );
    }
  }

  const arrowStyle = spotlight
    ? placement === "top" || placement === "bottom"
      ? {
          left: clamp((spotlight.left + spotlight.right) / 2 - popoverLeft, 24, popoverWidth - 24),
          [placement === "bottom" ? "top" : "bottom"]: -7,
          transform: "translateX(-50%) rotate(45deg)",
        }
      : {
          top: clamp((spotlight.top + spotlight.bottom) / 2 - popoverTop, 24, popoverHeight - 24),
          [placement === "right" ? "left" : "right"]: -7,
          transform: "translateY(-50%) rotate(45deg)",
        }
    : undefined;

  return createPortal(
    <div className="pointer-events-none fixed inset-0 z-[90]">
      {spotlight ? (
        <>
          <div
            className="pointer-events-auto absolute inset-x-0 top-0 bg-black/50"
            style={{ height: spotlight.top }}
          />
          <div
            className="pointer-events-auto absolute left-0 bg-black/50"
            style={{
              top: spotlight.top,
              width: spotlight.left,
              height: spotlight.bottom - spotlight.top,
            }}
          />
          <div
            className="pointer-events-auto absolute right-0 bg-black/50"
            style={{
              top: spotlight.top,
              width: viewport.width - spotlight.right,
              height: spotlight.bottom - spotlight.top,
            }}
          />
          <div
            className="pointer-events-auto absolute inset-x-0 bottom-0 bg-black/50"
            style={{ top: spotlight.bottom }}
          />
          <div
            className="pointer-events-auto absolute rounded-2xl border-2 border-brand-400 shadow-[0_0_0_4px_hsl(248_68%_61%_/_0.2)]"
            style={{
              left: spotlight.left,
              top: spotlight.top,
              width: spotlight.right - spotlight.left,
              height: spotlight.bottom - spotlight.top,
            }}
          />
        </>
      ) : (
        <div className="pointer-events-auto absolute inset-0 bg-black/50" />
      )}

      <dialog
        ref={popoverRef}
        open
        aria-modal="true"
        aria-labelledby="tour-title"
        aria-describedby="tour-description"
        className={[
          "pointer-events-auto fixed m-0 rounded-2xl border border-border bg-surface p-4 text-text-primary shadow-modal sm:p-5",
          !targetReady && "invisible",
        ]
          .filter(Boolean)
          .join(" ")}
        style={{ left: popoverLeft, top: popoverTop, width: popoverWidth }}
      >
        {spotlight && (
          <span
            aria-hidden
            className={[
              "absolute size-3.5 bg-surface",
              placement === "bottom"
                ? "border-l border-t border-border"
                : placement === "top"
                  ? "border-r border-b border-border"
                  : placement === "right"
                    ? "border-l border-b border-border"
                    : "border-r border-t border-border",
            ].join(" ")}
            style={arrowStyle}
          />
        )}
        <div className="relative flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 text-xs text-text-tertiary">
            <span className="grid size-7 place-items-center rounded-full bg-brand-600 font-semibold text-white">
              {step + 1}
            </span>
            <span>
              {step + 1} of {STEPS.length}
            </span>
          </div>
          <button
            type="button"
            onClick={onFinish}
            aria-label="Close guide"
            className="grid size-7 place-items-center rounded-full text-text-tertiary hover:bg-canvas-subtle hover:text-text-primary focus-visible:outline-2 focus-visible:outline-brand-500"
          >
            <X size={15} />
          </button>
        </div>
        <h2 id="tour-title" className="relative mt-3 text-lg font-semibold text-text-primary">
          {activeStep.title}
        </h2>
        <p id="tour-description" className="relative mt-2 text-sm leading-6 text-text-secondary">
          {activeStep.description}
        </p>
        <div className="relative mt-5 flex items-center justify-end gap-2">
          <Button variant="ghost" size="sm" onClick={onFinish}>
            Stop
          </Button>
          {step > 0 && (
            <Button
              variant="secondary"
              size="sm"
              onClick={() => onStepChange(step - 1)}
              iconLeft={<ArrowLeft size={14} />}
            >
              Previous
            </Button>
          )}
          <Button
            ref={nextRef}
            size="sm"
            disabled={!targetReady}
            onClick={step === STEPS.length - 1 ? onFinish : () => onStepChange(step + 1)}
            iconRight={step === STEPS.length - 1 ? undefined : <ArrowRight size={14} />}
          >
            {step === STEPS.length - 1 ? "Finish" : "Next"}
          </Button>
        </div>
      </dialog>
    </div>,
    document.body,
  );
}
