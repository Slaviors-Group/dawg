/**
 * Textarea — multi-line text input with label, hint, error.
 */
import { type TextareaHTMLAttributes, forwardRef, useId } from "react";

interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string;
  hint?: string;
  error?: string;
  wrapperClassName?: string;
}

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  (
    {
      label,
      hint,
      error,
      id,
      className = "",
      wrapperClassName = "",
      rows = 5,
      ...props
    },
    ref
  ) => {
    const generatedId = useId();
    const inputId = id ?? generatedId;

    return (
      <div className={["flex flex-col gap-1", wrapperClassName].join(" ")}>
        {label && (
          <label
            htmlFor={inputId}
            className="text-xs font-medium text-text-secondary"
          >
            {label}
          </label>
        )}
        <textarea
          ref={ref}
          id={inputId}
          rows={rows}
          className={[
            "w-full px-3 py-2 text-sm rounded-md font-[--font-mono]",
            "bg-canvas-subtle text-text-primary",
            "border transition-colors duration-[--duration-fast]",
            "placeholder:text-text-disabled resize-y leading-relaxed",
            "focus:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/50",
            "disabled:opacity-50 disabled:cursor-not-allowed",
            error
              ? "border-error-border"
              : "border-border focus:border-brand-400",
            className,
          ].join(" ")}
          {...props}
        />
        {error && (
          <p className="text-xs text-error-text" role="alert">
            {error}
          </p>
        )}
        {!error && hint && (
          <p className="text-xs text-text-tertiary">{hint}</p>
        )}
      </div>
    );
  }
);
Textarea.displayName = "Textarea";
