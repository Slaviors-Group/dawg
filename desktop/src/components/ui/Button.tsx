/**
 * Button — primary interaction primitive.
 *
 * Variants: primary | secondary | ghost | danger
 * Sizes: sm | md | lg
 * States: default | loading | disabled
 */
import { type ButtonHTMLAttributes, forwardRef } from "react";
import { CircleNotch } from "@phosphor-icons/react";

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
    "bg-brand-500 text-white hover:bg-brand-600 active:bg-brand-700 shadow-sm",
  secondary:
    "bg-surface text-text-primary border border-border hover:bg-surface-hover active:bg-surface-active shadow-sm",
  ghost:
    "bg-transparent text-text-secondary hover:bg-surface-hover hover:text-text-primary active:bg-surface-active",
  danger:
    "bg-error-bg text-error-text border border-error-border hover:brightness-95 active:brightness-90",
};

const sizeClasses: Record<Size, string> = {
  sm: "h-7 px-3 text-xs gap-1.5 rounded-sm",
  md: "h-9 px-4 text-sm gap-2 rounded-md",
  lg: "h-11 px-5 text-sm gap-2.5 rounded-md",
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
    ref
  ) => {
    const isDisabled = disabled || loading;
    return (
      <button
        ref={ref}
        type="button"
        disabled={isDisabled}
        className={[
          "inline-flex items-center justify-center font-medium whitespace-nowrap",
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
        {!loading && iconRight && (
          <span className="shrink-0">{iconRight}</span>
        )}
      </button>
    );
  }
);
Button.displayName = "Button";
