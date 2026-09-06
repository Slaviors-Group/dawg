import { useState, useRef, useEffect, useId } from "react";
import { CaretDown, Check } from "@phosphor-icons/react";
import { AnimatePresence, motion } from "framer-motion";

export interface SelectOption {
  value: string;
  label: string;
}

interface SelectProps {
  id?: string;
  label?: string;
  hint?: string;
  error?: string;
  className?: string;
  wrapperClassName?: string;
  options: SelectOption[];
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;
}

export const Select = ({
  id,
  label,
  hint,
  error,
  className = "",
  wrapperClassName = "",
  options,
  value,
  onChange,
  placeholder = "Select an option...",
  disabled = false,
}: SelectProps) => {
  const generatedId = useId();
  const selectId = id ?? generatedId;
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const selectedOption = options.find((opt) => opt.value === value);

  // Handle click outside to close dropdown
  useEffect(() => {
    const handleOutsideClick = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener("mousedown", handleOutsideClick);
    }
    return () => document.removeEventListener("mousedown", handleOutsideClick);
  }, [isOpen]);

  return (
    <div className={["flex flex-col gap-1 relative", wrapperClassName].join(" ")} ref={containerRef}>
      {label && (
        <label htmlFor={selectId} className="text-xs font-medium text-text-secondary">
          {label}
        </label>
      )}
      
      <button
        type="button"
        id={selectId}
        disabled={disabled}
        onClick={() => !disabled && setIsOpen(!isOpen)}
        className={[
          "w-full h-9 pl-3 pr-8 text-sm rounded-md text-left flex items-center justify-between",
          "bg-canvas-subtle",
          "border transition-colors duration-[--duration-fast] cursor-pointer",
          "focus:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/50",
          "disabled:opacity-50 disabled:cursor-not-allowed",
          error ? "border-error-border" : isOpen ? "border-brand-400" : "border-border hover:border-text-tertiary",
          !selectedOption ? "text-text-tertiary" : "text-text-primary",
          className,
        ].join(" ")}
      >
        <span className="truncate">{selectedOption ? selectedOption.label : placeholder}</span>
        <CaretDown
          size={12}
          weight="bold"
          className={["text-text-tertiary transition-transform duration-200", isOpen ? "rotate-180" : ""].join(" ")}
          aria-hidden
        />
      </button>

      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0, y: -4, scale: 0.98 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -4, scale: 0.98 }}
            transition={{ duration: 0.15, ease: "easeOut" }}
            className="absolute z-50 w-full mt-1 top-full bg-surface border border-border rounded-lg shadow-dropdown overflow-hidden"
          >
            <ul className="max-h-60 overflow-y-auto py-1 outline-none no-scrollbar">
              {options.length === 0 ? (
                <li className="px-3 py-2 text-sm text-text-tertiary text-center italic">
                  No options available
                </li>
              ) : (
                options.map((option) => {
                  const isSelected = option.value === value;
                  return (
                    <li
                      key={option.value}
                      onClick={() => {
                        onChange(option.value);
                        setIsOpen(false);
                      }}
                      className={[
                        "flex items-center justify-between px-3 py-2 text-sm cursor-pointer transition-colors duration-fast",
                        isSelected ? "bg-brand-500/10 text-brand-600 font-medium" : "text-text-primary hover:bg-surface-hover",
                      ].join(" ")}
                    >
                      <span className="truncate">{option.label}</span>
                      {isSelected && <Check size={14} weight="bold" className="text-brand-500 shrink-0 ml-2" />}
                    </li>
                  );
                })
              )}
            </ul>
          </motion.div>
        )}
      </AnimatePresence>

      {error && (
        <p className="text-xs text-error-text mt-0.5" role="alert">
          {error}
        </p>
      )}
      {!error && hint && (
        <p className="text-xs text-text-tertiary mt-0.5">{hint}</p>
      )}
    </div>
  );
};
