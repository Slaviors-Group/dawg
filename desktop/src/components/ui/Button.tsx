import { CircleNotch } from "@phosphor-icons/react";

import { type ButtonHTMLAttributes, forwardRef } from "react";

type Variant = "primary" | "secondary" | "ghost" | "danger";
type Size = "sm" | "md" | "lg";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  loading?: boolean;
  iconLeft?: React.ReactNode;
  iconRight?: React.ReactNode;
}

const variantClasses: Record<Variant, string> = {
  primary:
    "bg-brand-600 text-white hover:bg-[hsl(248_65%_46%)] active:bg-[hsl(248_65%_40%)] shadow-[0_7px_16px_-8px_var(--color-brand-600)]",
  secondary:
    "bg-surface text-text-primary border border-border hover:border-brand-300 hover:bg-surface-hover active:bg-surface-active",
  ghost:
    "bg-transparent text-text-secondary hover:bg-brand-50 hover:text-brand-700 active:bg-brand-100",
  danger:
    "bg-error-bg text-error-text border border-error-border hover:brightness-95 active:brightness-90",
};

const sizeClasses: Record<Size, string> = {
  sm: "min-h-8 px-3.5 text-xs gap-1.5 rounded-full",
  md: "min-h-10 px-4.5 text-sm gap-2 rounded-full",
  lg: "min-h-12 px-6 text-sm gap-2.5 rounded-full",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      variant = "primary",
      size = "md",
      loading = false,
      iconLeft,
      iconRight,
      className = "",
      disabled,
      children,
      ...props
    },
    ref,
  ) => {
    const isDisabled = disabled || loading;
    return (
      <button
        ref={ref}
        type="button"
        disabled={isDisabled}
        className={[
          "inline-flex items-center justify-center font-semibold whitespace-nowrap",
          "transition-all duration-[--duration-fast] select-none",
          "focus-visible:outline-2 focus-visible:outline-[--color-brand-500] focus-visible:outline-offset-2",
          "disabled:opacity-50 disabled:pointer-events-none",
          variantClasses[variant],
          sizeClasses[size],
          className,
        ].join(" ")}
        {...props}
      >
        {loading ? (
          <CircleNotch
            size={size === "sm" ? 12 : 14}
            className="animate-spin shrink-0"
            aria-hidden
          />
        ) : (
          iconLeft && <span className="shrink-0">{iconLeft}</span>
        )}
        {children}
        {!loading && iconRight && <span className="shrink-0">{iconRight}</span>}
      </button>
    );
  },
);
Button.displayName = "Button";
