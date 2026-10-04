import type { ReactNode } from "react";

interface StatCardProps {
  label: string;
  value: ReactNode;
  icon?: ReactNode;
  badge?: ReactNode;
  trend?: { direction: "up" | "down" | "neutral"; label: string };
  className?: string;
}

export function StatCard({ label, value, icon, badge, trend, className = "" }: StatCardProps) {
  const trendColor =
    trend?.direction === "up"
      ? "text-success-text"
      : trend?.direction === "down"
        ? "text-error-text"
        : "text-text-tertiary";

  return (
    <div
      className={[
        "bg-surface border border-border/70 rounded-2xl",
        "p-5 flex flex-col gap-4 shadow-card",
        className,
      ].join(" ")}
    >
      <div className="flex items-center justify-between gap-2">
        <span className="text-xs font-semibold text-text-secondary">{label}</span>
        {icon && (
          <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-brand-100 text-brand-700">
            {icon}
          </span>
        )}
      </div>
      <div className="flex items-end justify-between gap-2">
        <span className="text-[1.75rem] font-semibold text-text-primary leading-none tracking-tight">
          {value}
        </span>
        {badge}
      </div>
      {trend && <p className={`text-xs font-medium ${trendColor}`}>{trend.label}</p>}
    </div>
  );
}
