import { motion } from "framer-motion";
import type { ReactNode } from "react";

interface PageShellProps {
  title: string;
  /** A second part of the title highlighted in the brand color. */
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
      className={["flex min-h-full flex-col gap-6", className].filter(Boolean).join(" ")}
      initial={{ opacity: 0, y: 6 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2, ease: "easeOut" }}
    >
      {/* Page header */}
      <div className="flex flex-wrap items-end justify-between gap-4 shrink-0">
        <div>
          {eyebrow && (
            <span className="text-[11px] font-semibold text-brand-600 uppercase tracking-[0.16em] block mb-2">
              {eyebrow}
            </span>
          )}
          <h1 className="text-[2rem] font-semibold text-text-primary leading-tight tracking-tight sm:text-[2.25rem]">
            {title} {titleAccent && <span className="text-brand-500">{titleAccent}</span>}
          </h1>
          {subtitle && (
            <p className="text-sm text-text-secondary mt-1.5 max-w-2xl leading-relaxed">
              {subtitle}
            </p>
          )}
        </div>
        {actions && <div className="flex items-center gap-2 shrink-0">{actions}</div>}
      </div>

      {/* Page body */}
      <div className={["flex flex-col gap-5", bodyClassName].filter(Boolean).join(" ")}>
        {children}
      </div>
    </motion.div>
  );
}
