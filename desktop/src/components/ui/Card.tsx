import { type HTMLAttributes, forwardRef } from "react";

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  accent?: boolean;

  noPad?: boolean;
}

export const Card = forwardRef<HTMLDivElement, CardProps>(
  ({ accent = false, noPad = false, className = "", children, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={[
          "bg-surface border border-border/70 rounded-2xl",
          "shadow-card",
          !noPad && "p-5 sm:p-6",
          accent && "border-l-4 border-l-brand-500",
          className,
        ]
          .filter(Boolean)
          .join(" ")}
        {...props}
      >
        {children}
      </div>
    );
  },
);
Card.displayName = "Card";

export function CardHeader({
  className = "",
  children,
  bordered = false,
  ...props
}: HTMLAttributes<HTMLDivElement> & { bordered?: boolean }) {
  return (
    <div
      className={[
        "flex items-center justify-between gap-4",
        bordered ? "pb-4 mb-4 border-b border-border" : "mb-5",
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

export function CardTitle({
  title,
  subtitle,
}: {
  title: string;
  subtitle?: string;
}) {
  return (
    <div>
      <h2 className="text-[17px] font-semibold text-text-primary leading-tight tracking-tight">
        {title}
      </h2>
      {subtitle && <p className="text-xs text-text-tertiary mt-1">{subtitle}</p>}
    </div>
  );
}
