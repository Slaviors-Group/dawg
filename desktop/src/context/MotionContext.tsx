import {
  type ReactNode,
  createContext,
  useContext,
  useEffect,
  useState,
} from "react";
import { MotionConfig } from "framer-motion";

interface MotionContextType {
  /** true = animations enabled, false = reduced / disabled */
  animationsEnabled: boolean;
  setAnimationsEnabled: (v: boolean) => void;
}

const MotionContext = createContext<MotionContextType | undefined>(undefined);

export function MotionProvider({ children }: { children: ReactNode }) {
  const [animationsEnabled, setAnimationsState] = useState<boolean>(() => {
    // Honor OS preference first, then stored setting
    const systemReduced = window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;
    if (systemReduced) return false;
    const stored = localStorage.getItem("dawg-animations");
    return stored !== null ? stored === "true" : true;
  });

  // Apply data-motion="reduced" attribute to <html> when disabled
  useEffect(() => {
    const html = document.documentElement;
    if (!animationsEnabled) {
      html.setAttribute("data-motion", "reduced");
    } else {
      html.removeAttribute("data-motion");
    }
  }, [animationsEnabled]);

  const setAnimationsEnabled = (v: boolean) => {
    setAnimationsState(v);
    localStorage.setItem("dawg-animations", String(v));
  };

  return (
    <MotionContext.Provider value={{ animationsEnabled, setAnimationsEnabled }}>
      <MotionConfig reducedMotion={animationsEnabled ? "never" : "always"}>
        {children}
      </MotionConfig>
    </MotionContext.Provider>
  );
}

export function useMotion() {
  const ctx = useContext(MotionContext);
  if (!ctx) throw new Error("useMotion must be used within MotionProvider");
  return ctx;
}
