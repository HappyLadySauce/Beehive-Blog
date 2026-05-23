"use client";

import { createContext, type ReactNode, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { Monitor, Moon, Sun } from "lucide-react";

type ThemePreference = "system" | "light" | "dark";
type ResolvedTheme = "light" | "dark";

type ThemeContextValue = {
  preference: ThemePreference;
  resolvedTheme: ResolvedTheme;
  cycleTheme: () => void;
  setPreference: (preference: ThemePreference) => void;
};

const storageKey = "beehive.theme";
const ThemeContext = createContext<ThemeContextValue | null>(null);

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [preference, setPreferenceState] = useState<ThemePreference>(() => {
    if (typeof window === "undefined") return "system";
    const stored = readStoredPreference();
    return stored === "system" || stored === "light" || stored === "dark" ? stored : "system";
  });
  const [systemTheme, setSystemTheme] = useState<ResolvedTheme>(() => {
    if (typeof window === "undefined") return "light";
    return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
  });

  useEffect(() => {
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    const syncSystemTheme = () => setSystemTheme(media.matches ? "dark" : "light");
    media.addEventListener("change", syncSystemTheme);
    return () => media.removeEventListener("change", syncSystemTheme);
  }, []);

  const resolvedTheme: ResolvedTheme = preference === "system" ? systemTheme : preference;

  useEffect(() => {
    document.documentElement.dataset.theme = resolvedTheme;
  }, [resolvedTheme]);

  const setPreference = useCallback((next: ThemePreference) => {
    setPreferenceState(next);
    writeStoredPreference(next);
  }, []);

  const cycleTheme = useCallback(() => {
    const next: ThemePreference = preference === "system" ? "light" : preference === "light" ? "dark" : "system";
    setPreference(next);
  }, [preference, setPreference]);

  const value = useMemo(
    () => ({ cycleTheme, preference, resolvedTheme, setPreference }),
    [cycleTheme, preference, resolvedTheme, setPreference]
  );

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

function readStoredPreference() {
  try {
    return window.localStorage?.getItem(storageKey) ?? null;
  } catch {
    return null;
  }
}

function writeStoredPreference(next: ThemePreference) {
  try {
    window.localStorage?.setItem(storageKey, next);
  } catch {
    // Storage can be unavailable in restricted browser contexts.
    // 受限浏览器上下文中存储可能不可用。
  }
}

export function useTheme() {
  const value = useContext(ThemeContext);
  if (!value) {
    throw new Error("useTheme must be used within ThemeProvider");
  }
  return value;
}

export function ThemeToggleButton({ className }: { className?: string }) {
  const fallback = useContext(ThemeContext);
  const preference = fallback?.preference ?? "system";
  const resolvedTheme = fallback?.resolvedTheme ?? "light";
  const cycleTheme = fallback?.cycleTheme ?? (() => undefined);
  const label = `切换显示模式，当前为${preference === "system" ? `跟随系统（${resolvedTheme === "dark" ? "深色" : "浅色"}）` : preference === "dark" ? "深色" : "浅色"}`;

  return (
    <button
      className={className ?? "theme-toggle-button"}
      type="button"
      aria-label={label}
      title={label}
      data-theme-toggle="true"
      onClick={cycleTheme}
    >
      {preference === "system" ? <Monitor aria-hidden size={18} /> : resolvedTheme === "dark" ? <Moon aria-hidden size={18} /> : <Sun aria-hidden size={18} />}
    </button>
  );
}
