/**
 * Separator — horizontal rule with optional label.
 */
interface SeparatorProps {
  label?: string;
  className?: string;
}

export function Separator({ label, className = "" }: SeparatorProps) {
  if (label) {
    return (
      <div className={["flex items-center gap-3", className].join(" ")}>
        <div className="flex-1 h-px bg-border" />
        <span className="text-[10px] font-medium text-text-tertiary uppercase tracking-wider whitespace-nowrap">
          {label}
        </span>
        <div className="flex-1 h-px bg-border" />
      </div>
    );
  }
  return <div className={["h-px bg-border", className].join(" ")} aria-hidden />;
}
