"use client";

import { createContext, type ReactNode, useCallback, useContext, useEffect, useMemo, useSyncExternalStore } from "react";
import { Monitor, Moon, Sun } from "lucide-react";

import type { ResolvedTheme, ThemePreference } from "@/lib/theme/constants";
import { resolveThemeFromPreference } from "@/lib/theme/resolve";
import {
  ensureThemePreferenceSynced,
  getThemeServerSnapshot,
  getThemeSnapshot,
  setThemePreference,
  subscribeThemeStore
} from "@/lib/theme/store";

type ThemeContextValue = {
  preference: ThemePreference;
  resolvedTheme: ResolvedTheme;
  cycleTheme: () => void;
  setPreference: (preference: ThemePreference) => void;
};

const ThemeContext = createContext<ThemeContextValue | null>(null);

export function ThemeProvider({ children }: { children: ReactNode }) {
  const snapshot = useSyncExternalStore(subscribeThemeStore, getThemeSnapshot, getThemeServerSnapshot);
  const { preference, systemTheme } = snapshot;
  const resolvedTheme = resolveThemeFromPreference(preference, systemTheme);

  useEffect(() => {
    ensureThemePreferenceSynced();
  }, []);

  useEffect(() => {
    document.documentElement.dataset.theme = resolvedTheme;
  }, [resolvedTheme]);

  const setPreference = useCallback((next: ThemePreference) => {
    setThemePreference(next);
  }, []);

  const cycleTheme = useCallback(() => {
    const next: ThemePreference = preference === "system" ? "light" : preference === "light" ? "dark" : "system";
    setThemePreference(next);
  }, [preference]);

  const value = useMemo(
    () => ({ cycleTheme, preference, resolvedTheme, setPreference }),
    [cycleTheme, preference, resolvedTheme, setPreference]
  );

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
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
