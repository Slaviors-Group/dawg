/**
 * Badge — compact semantic label.
 *
 * Variants: default | success | warning | error | info | brand
 * Sizes: sm | md
 * Optional dot indicator prefix.
 */

type Variant = "default" | "success" | "warning" | "error" | "info" | "brand";
type Size = "sm" | "md";

interface BadgeProps {
  variant?: Variant;
  size?: Size;
  dot?: boolean;
  children: React.ReactNode;
  className?: string;
}

const variantClasses: Record<Variant, string> = {
  default:
    "bg-canvas-subtle text-text-secondary border-border",
  brand:
    "bg-brand-100 text-brand-700 border-brand-200",
  success:
    "bg-success-bg text-success-text border-success-border",
  warning:
    "bg-warning-bg text-warning-text border-warning-border",
  error:
    "bg-error-bg text-error-text border-error-border",
  info:
    "bg-info-bg text-info-text border-info-border",
};

const dotClasses: Record<Variant, string> = {
  default: "bg-text-tertiary",
  brand:   "bg-brand-500",
  success: "bg-success-dot",
  warning: "bg-warning-dot",
  error:   "bg-error-dot",
  info:    "bg-info-dot",
};

const sizeClasses: Record<Size, string> = {
  sm: "px-1.5 py-0.5 text-[10px] gap-1 rounded-sm",
  md: "px-2 py-0.5 text-xs gap-1.5 rounded-sm",
};

export function Badge({
  variant = "default",
  size = "md",
  dot = false,
  children,
  className = "",
}: BadgeProps) {
  return (
    <span
      className={[
        "inline-flex items-center font-medium border",
        variantClasses[variant],
        sizeClasses[size],
        className,
      ].join(" ")}
    >
      {dot && (
        <span
          className={[
            "shrink-0 rounded-full",
            size === "sm" ? "w-1 h-1" : "w-1.5 h-1.5",
            dotClasses[variant],
          ].join(" ")}
          aria-hidden
        />
      )}
      {children}
    </span>
  );
}
