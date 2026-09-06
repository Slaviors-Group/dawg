/**
 * StatCard — dashboard metric tile.
 * Shows a label, large numeric/text value, optional badge and trend indicator.
 */
import type { ReactNode } from "react";

interface StatCardProps {
  label: string;
  value: ReactNode;
  icon?: ReactNode;
  badge?: ReactNode;
  trend?: { direction: "up" | "down" | "neutral"; label: string };
  className?: string;
}

export function StatCard({
  label,
  value,
  icon,
  badge,
  trend,
  className = "",
}: StatCardProps) {
  const trendColor =
    trend?.direction === "up"
      ? "text-success-text"
      : trend?.direction === "down"
        ? "text-error-text"
        : "text-text-tertiary";

  return (
    <div
      className={[
        "bg-surface border border-border rounded-lg",
        "p-4 flex flex-col gap-3 shadow-card",
        className,
      ].join(" ")}
    >
      <div className="flex items-center justify-between gap-2">
        <span className="text-xs font-medium text-text-secondary uppercase tracking-wide">
          {label}
        </span>
        {icon && (
          <span className="text-text-tertiary shrink-0">{icon}</span>
        )}
      </div>
      <div className="flex items-end justify-between gap-2">
        <span className="text-2xl font-semibold text-text-primary leading-none">
          {value}
        </span>
        {badge}
      </div>
      {trend && (
        <p className={`text-xs font-medium ${trendColor}`}>{trend.label}</p>
      )}
    </div>
  );
}
