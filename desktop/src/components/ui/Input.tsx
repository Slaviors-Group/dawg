/**
 * Input — single-line text/url/number input with label, hint, and error states.
 */
import { type InputHTMLAttributes, forwardRef, useId } from "react";

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  hint?: string;
  error?: string;
  /** Phosphor icon to render inside the left edge */
  iconLeft?: React.ReactNode;
  /** Phosphor icon or node to render inside the right edge */
  iconRight?: React.ReactNode;
  /** Wrapper className */
  wrapperClassName?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  (
    {
      label,
      hint,
      error,
      iconLeft,
      iconRight,
      id,
      className = "",
      wrapperClassName = "",
      disabled,
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
        <div className="relative flex items-center">
          {iconLeft && (
            <span
              className="absolute left-2.5 text-text-tertiary pointer-events-none"
              aria-hidden
            >
              {iconLeft}
            </span>
          )}
          <input
            ref={ref}
            id={inputId}
            disabled={disabled}
            className={[
              "w-full h-9 px-3 text-sm rounded-md font-[--font-sans]",
              "bg-canvas-subtle text-text-primary",
              "border transition-colors duration-[--duration-fast]",
              "placeholder:text-text-disabled",
              "focus:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/50",
              "disabled:opacity-50 disabled:cursor-not-allowed",
              error
                ? "border-error-border focus:border-error-border"
                : "border-border focus:border-brand-400",
              iconLeft && "pl-9",
              iconRight && "pr-9",
              className,
            ].join(" ")}
            {...props}
          />
          {iconRight && (
            <span
              className="absolute right-2.5 text-text-tertiary"
              aria-hidden
            >
              {iconRight}
            </span>
          )}
        </div>
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
Input.displayName = "Input";
