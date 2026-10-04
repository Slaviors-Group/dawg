import { type InputHTMLAttributes, forwardRef, useId } from "react";

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  hint?: string;
  error?: string;

  iconLeft?: React.ReactNode;

  iconRight?: React.ReactNode;

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
    ref,
  ) => {
    const generatedId = useId();
    const inputId = id ?? generatedId;

    return (
      <div className={["flex flex-col gap-1.5", wrapperClassName].join(" ")}>
        {label && (
          <label htmlFor={inputId} className="text-xs font-semibold text-text-secondary">
            {label}
          </label>
        )}
        <div className="relative flex items-center">
          {iconLeft && (
            <span className="absolute left-3.5 text-text-tertiary pointer-events-none" aria-hidden>
              {iconLeft}
            </span>
          )}
          <input
            ref={ref}
            id={inputId}
            disabled={disabled}
            className={[
              "w-full h-11 px-4 text-sm rounded-xl font-[--font-sans]",
              "bg-surface text-text-primary shadow-[0_1px_2px_hsl(240_20%_20%_/_0.02)]",
              "border transition-colors duration-[--duration-fast]",
              "placeholder:text-text-disabled",
              "focus:outline-none focus-visible:ring-4 focus-visible:ring-brand-500/10",
              "disabled:opacity-50 disabled:cursor-not-allowed",
              error
                ? "border-error-border focus:border-error-border"
                : "border-border focus:border-brand-400",
              iconLeft && "pl-10",
              iconRight && "pr-10",
              className,
            ].join(" ")}
            {...props}
          />
          {iconRight && (
            <span className="absolute right-3.5 text-text-tertiary" aria-hidden>
              {iconRight}
            </span>
          )}
        </div>
        {error && (
          <p className="text-xs text-error-text" role="alert">
            {error}
          </p>
        )}
        {!error && hint && <p className="text-xs text-text-tertiary">{hint}</p>}
      </div>
    );
  },
);
Input.displayName = "Input";
