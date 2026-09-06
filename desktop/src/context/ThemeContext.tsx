import {
  type ReactNode,
  createContext,
  useContext,
  useEffect,
  useState,
} from "react";
import { flushSync } from "react-dom";

export type ThemeValue = "light" | "dark" | "system";

interface ThemeContextType {
  theme: ThemeValue;
  resolvedTheme: "light" | "dark";
  setTheme: (t: ThemeValue) => void;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

function getSystemTheme(): "light" | "dark" {
  if (typeof window === "undefined") return "light";
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<ThemeValue>(() => {
    const saved = localStorage.getItem("dawg-theme") as ThemeValue | null;
    return saved ?? "light";
  });

  const resolvedTheme: "light" | "dark" =
    theme === "system" ? getSystemTheme() : theme;

  // Apply .dark class to <html> and listen for system changes
  useEffect(() => {
    const html = document.documentElement;
    const apply = (resolved: "light" | "dark") => {
      if (resolved === "dark") {
        html.classList.add("dark");
      } else {
        html.classList.remove("dark");
      }
    };

    if (theme === "system") {
      const mq = window.matchMedia("(prefers-color-scheme: dark)");
      apply(mq.matches ? "dark" : "light");
      const listener = (e: MediaQueryListEvent) =>
        apply(e.matches ? "dark" : "light");
      mq.addEventListener("change", listener);
      return () => mq.removeEventListener("change", listener);
    }
    apply(theme);
  }, [theme]);

  const setTheme = (t: ThemeValue) => {
    const applyChanges = () => {
      flushSync(() => {
        setThemeState(t);
      });
      // Force DOM class update synchronously for View Transition API
      const resolved = t === "system" ? getSystemTheme() : t;
      if (resolved === "dark") {
        document.documentElement.classList.add("dark");
      } else {
        document.documentElement.classList.remove("dark");
      }
    };

    const prefersReducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    
    // Check if API is supported and motion is allowed
    if (!document.startViewTransition || prefersReducedMotion) {
      applyChanges();
    } else {
      document.startViewTransition(() => {
        applyChanges();
      });
    }
    
    localStorage.setItem("dawg-theme", t);
  };

  return (
    <ThemeContext.Provider value={{ theme, resolvedTheme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error("useTheme must be used within ThemeProvider");
  return ctx;
}
