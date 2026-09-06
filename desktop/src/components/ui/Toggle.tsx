/**
 * Toggle (Switch) — binary on/off control with label.
 * Used for settings like animations enabled, etc.
 */
import { useId } from "react";

interface ToggleProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  label?: string;
  hint?: string;
  disabled?: boolean;
  id?: string;
}

export function Toggle({
  checked,
  onChange,
  label,
  hint,
  disabled = false,
  id,
}: ToggleProps) {
  const generatedId = useId();
  const toggleId = id ?? generatedId;

  return (
    <div className="flex items-center gap-3">
      <button
        id={toggleId}
        role="switch"
        type="button"
        aria-checked={checked}
        disabled={disabled}
        onClick={() => onChange(!checked)}
        className={[
          "relative inline-flex items-center w-9 h-5 rounded-full",
          "transition-colors duration-[--duration-base] shrink-0",
          "focus-visible:outline-2 focus-visible:outline-[--color-brand-500] focus-visible:outline-offset-2",
          "disabled:opacity-50 disabled:cursor-not-allowed",
          checked
            ? "bg-brand-500"
            : "bg-border-strong",
        ].join(" ")}
      >
        <span
          className={[
            "absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full shadow-sm",
            "transition-transform duration-[--duration-base]",
            checked ? "translate-x-4" : "translate-x-0",
          ].join(" ")}
        />
      </button>
      {(label || hint) && (
        <label htmlFor={toggleId} className="cursor-pointer">
          {label && (
            <span className="text-sm font-medium text-text-primary block">
              {label}
            </span>
          )}
          {hint && (
            <span className="text-xs text-text-tertiary block">
              {hint}
            </span>
          )}
        </label>
      )}
    </div>
  );
}
