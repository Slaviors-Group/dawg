import { MonitorPlay, Moon, Sun } from "@phosphor-icons/react";
import { useMotion } from "../../context/MotionContext";
/**
 * SettingsPanel — slide-in settings drawer triggered from the sidebar.
 * Contains theme selector, animations toggle, and app info.
 */
import { type ThemeValue, useTheme } from "../../context/ThemeContext";
import { Modal } from "./Modal";
import { Separator } from "./Separator";
import { Toggle } from "./Toggle";

interface SettingsPanelProps {
  open: boolean;
  onClose: () => void;
}

const THEME_OPTIONS: { value: ThemeValue; label: string; icon: typeof Sun }[] = [
  { value: "light", label: "Light", icon: Sun },
  { value: "dark", label: "Dark", icon: Moon },
  { value: "system", label: "System", icon: MonitorPlay },
];

export function SettingsPanel({ open, onClose }: SettingsPanelProps) {
  const { theme, setTheme } = useTheme();
  const { animationsEnabled, setAnimationsEnabled } = useMotion();

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Settings"
      subtitle="Appearance and accessibility preferences"
      maxWidth="sm"
    >
      <div className="flex flex-col gap-5">
        {/* Theme */}
        <section className="flex flex-col gap-3">
          <h4 className="text-xs font-semibold text-text-tertiary uppercase tracking-wider">
            Appearance
          </h4>
          <div className="grid grid-cols-3 gap-2">
            {THEME_OPTIONS.map(({ value, label, icon: Icon }) => {
              const isActive = theme === value;
              return (
                <button
                  key={value}
                  type="button"
                  onClick={() => setTheme(value)}
                  className={[
                    "flex flex-col items-center gap-2 p-3 rounded-md border text-sm font-medium",
                    "transition-all duration-[--duration-fast]",
                    isActive
                      ? "bg-brand-100 border-brand-400 text-brand-700"
                      : "bg-canvas-subtle border-border text-text-secondary",
                    "hover:border-brand-300",
                  ].join(" ")}
                  aria-pressed={isActive}
                >
                  <Icon size={20} weight={isActive ? "fill" : "regular"} />
                  {label}
                </button>
              );
            })}
          </div>
        </section>

        <Separator />

        {/* Animations */}
        <section className="flex flex-col gap-3">
          <h4 className="text-xs font-semibold text-text-tertiary uppercase tracking-wider">
            Accessibility
          </h4>
          <Toggle
            checked={animationsEnabled}
            onChange={setAnimationsEnabled}
            label="Interface animations"
            hint="Disable for reduced visual motion."
          />
        </section>

        <Separator />

        {/* About */}
        <section className="flex flex-col gap-1">
          <h4 className="text-xs font-semibold text-text-tertiary uppercase tracking-wider mb-1">
            About
          </h4>
          <div className="flex justify-between text-xs text-text-tertiary">
            <span>DAWG Desktop Shell</span>
            <span className="font-mono">v0.1.3-alpha</span>
          </div>
          <div className="flex justify-between text-xs text-text-tertiary mt-1">
            <span>Developer</span>
            <span className="font-mono">Slaviors</span>
          </div>
        </section>
      </div>
    </Modal>
  );
}
