/**
 * ProgressBar — linear progress indicator.
 * Used to show capture session duration, replay progress, etc.
 */
interface ProgressBarProps {
  /** 0–100 */
  value: number;
  /** Animate the fill? */
  animated?: boolean;
  label?: string;
  size?: "sm" | "md";
  variant?: "brand" | "success" | "warning" | "error";
}

const trackHeight = { sm: "h-1", md: "h-2" };

const fillColor = {
  brand: "bg-brand-500",
  success: "bg-success-dot",
  warning: "bg-warning-dot",
  error: "bg-error-dot",
};

export function ProgressBar({
  value,
  animated = false,
  label,
  size = "md",
  variant = "brand",
}: ProgressBarProps) {
  const clampedValue = Math.min(100, Math.max(0, value));

  return (
    <div className="flex flex-col gap-1">
      {label && (
        <div className="flex items-center justify-between text-xs text-text-tertiary">
          <span>{label}</span>
          <span className="font-mono">{clampedValue}%</span>
        </div>
      )}
      <div
        role="progressbar"
        tabIndex={0}
        aria-valuenow={clampedValue}
        aria-valuemin={0}
        aria-valuemax={100}
        className={["w-full rounded-full bg-canvas-subtle overflow-hidden", trackHeight[size]].join(
          " ",
        )}
      >
        <div
          style={{ width: `${clampedValue}%` }}
          className={[
            "h-full rounded-full transition-[width] duration-[--duration-slow]",
            fillColor[variant],
            animated && "animate-pulse",
          ]
            .filter(Boolean)
            .join(" ")}
        />
      </div>
    </div>
  );
}
