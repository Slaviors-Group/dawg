import { motion } from "framer-motion";
import type { ReactNode } from "react";

interface PageShellProps {
  title: string;
  /** A second part of the title that is highlighted in a cyan/teal color */
  titleAccent?: string;
  /** A smaller description text rendered below the main title */
  subtitle?: string;
  /** A small eyebrow text above the main title (e.g., date or section context) */
  eyebrow?: string;
  /** Action buttons rendered to the right of the title */
  actions?: ReactNode;
  /** Optional classes for pages that manage their own fixed/scrollable regions. */
  className?: string;
  /** Optional classes applied to the page body wrapper. */
  bodyClassName?: string;
  children: ReactNode;
}

export function PageShell({
  title,
  titleAccent,
  subtitle,
  eyebrow,
  actions,
  className,
  bodyClassName,
  children,
}: PageShellProps) {
  return (
    <motion.div
      key={title}
      className={["flex min-h-full flex-col gap-8", className].filter(Boolean).join(" ")}
      initial={{ opacity: 0, y: 6 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2, ease: "easeOut" }}
    >
      {/* Page header */}
      <div className="flex items-start justify-between gap-4 shrink-0">
        <div>
          {eyebrow && (
            <span className="text-xs font-semibold text-text-tertiary uppercase tracking-wider block mb-3">
              {eyebrow}
            </span>
          )}
          <h1 className="text-[2.5rem] font-bold text-text-primary leading-none tracking-tight">
            {title} {titleAccent && <span className="text-brand-500">{titleAccent}</span>}
          </h1>
          {subtitle && (
            <p className="text-sm text-text-tertiary mt-3 max-w-2xl leading-relaxed">{subtitle}</p>
          )}
        </div>
        {actions && <div className="flex items-center gap-2 shrink-0">{actions}</div>}
      </div>

      {/* Page body */}
      <div className={["flex flex-col gap-6", bodyClassName].filter(Boolean).join(" ")}>
        {children}
      </div>
    </motion.div>
  );
}
