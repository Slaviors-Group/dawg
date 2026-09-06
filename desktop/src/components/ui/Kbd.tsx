/**
 * Kbd — keyboard shortcut display tag.
 */
interface KbdProps {
  children: React.ReactNode;
  className?: string;
}

export function Kbd({ children, className = "" }: KbdProps) {
  return (
    <kbd
      className={[
        "inline-flex items-center justify-center",
        "px-1.5 py-0.5 text-[10px] font-[--font-mono] font-medium",
        "bg-canvas-subtle border border-border rounded-sm",
        "text-text-tertiary shadow-[0_1px_0_0_var(--color-border)]",
        className,
      ].join(" ")}
    >
      {children}
    </kbd>
  );
}
