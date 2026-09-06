/**
 * Card — base surface container.
 * All card-like elements in the app are built on this.
 */
import { type HTMLAttributes, forwardRef } from "react";

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  /** Apply a thicker brand-coloured left border accent */
  accent?: boolean;
  /** Remove default padding */
  noPad?: boolean;
}

export const Card = forwardRef<HTMLDivElement, CardProps>(
  ({ accent = false, noPad = false, className = "", children, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={[
          "bg-surface border border-border rounded-lg",
          "shadow-card",
          !noPad && "p-5",
          accent && "border-l-2 border-l-[--color-brand-500]",
          className,
        ]
          .filter(Boolean)
          .join(" ")}
        {...props}
      >
        {children}
      </div>
    );
  }
);
Card.displayName = "Card";

/** Preset card header section with optional border-bottom */
export function CardHeader({
  className = "",
  children,
  bordered = true,
  ...props
}: HTMLAttributes<HTMLDivElement> & { bordered?: boolean }) {
  return (
    <div
      className={[
        "flex items-center justify-between gap-4",
        bordered && "pb-4 mb-4 border-b border-border",
        className,
      ]
        .filter(Boolean)
        .join(" ")}
      {...props}
    >
      {children}
    </div>
  );
}

/** Section heading + subtitle inside a Card */
export function CardTitle({
  title,
  subtitle,
}: {
  title: string;
  subtitle?: string;
}) {
  return (
    <div>
      <h2 className="text-base font-semibold text-text-primary leading-tight">
        {title}
      </h2>
      {subtitle && (
        <p className="text-xs text-text-tertiary mt-0.5">
          {subtitle}
        </p>
      )}
    </div>
  );
}
