/**
 * NavItem — sidebar navigation link button.
 * Shows icon + label with active/hover states.
 */
import type { ReactNode } from "react";

interface NavItemProps {
  icon: ReactNode;
  label: string;
  active?: boolean;
  badge?: ReactNode;
  onClick?: () => void;
}

export function NavItem({ icon, label, active = false, badge, onClick }: NavItemProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={[
        "w-full flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium",
        "transition-all duration-[--duration-fast] group",
        active
          ? "bg-brand-100 text-brand-700"
          : "text-text-secondary hover:bg-surface-hover hover:text-text-primary",
      ].join(" ")}
      aria-current={active ? "page" : undefined}
    >
      <span
        className={[
          "shrink-0 transition-colors duration-[--duration-fast]",
          active ? "text-brand-600" : "text-text-tertiary group-hover:text-text-secondary",
        ].join(" ")}
        aria-hidden
      >
        {icon}
      </span>
      <span className="flex-1 text-left leading-none">{label}</span>
      {badge && <span className="shrink-0">{badge}</span>}
    </button>
  );
}
