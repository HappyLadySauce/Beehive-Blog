import { secureCookieEnabled } from "@/lib/auth/cookies";

import {
  DEFAULT_SYSTEM_THEME,
  DEFAULT_THEME_PREFERENCE,
  isThemePreference,
  type ThemePreference,
  THEME_COOKIE_NAME,
  THEME_STORAGE_KEY
} from "./constants";
import { getServerThemeSnapshot, type ThemeSnapshot } from "./resolve";

export function readStoredThemePreference(): ThemePreference | null {
  if (typeof window === "undefined") {
    return null;
  }
  try {
    const stored = window.localStorage?.getItem(THEME_STORAGE_KEY) ?? readThemePreferenceFromDocumentCookie();
    return isThemePreference(stored) ? stored : null;
  } catch {
    return readThemePreferenceFromDocumentCookie();
  }
}

export function readThemePreferenceFromDocumentCookie(): ThemePreference | null {
  if (typeof document === "undefined") {
    return null;
  }
  const match = document.cookie.match(new RegExp(`(?:^|; )${THEME_COOKIE_NAME}=([^;]*)`));
  const value = match?.[1] ? decodeURIComponent(match[1]) : null;
  return isThemePreference(value) ? value : null;
}

export function readClientSystemTheme(): typeof DEFAULT_SYSTEM_THEME | "dark" {
  if (typeof window === "undefined") {
    return DEFAULT_SYSTEM_THEME;
  }
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

export function getClientThemeSnapshot(): ThemeSnapshot {
  return {
    preference: readStoredThemePreference() ?? DEFAULT_THEME_PREFERENCE,
    systemTheme: readClientSystemTheme()
  };
}

export function writeThemePreference(next: ThemePreference) {
  try {
    window.localStorage?.setItem(THEME_STORAGE_KEY, next);
  } catch {
    // Storage can be unavailable in restricted browser contexts.
    // 受限浏览器上下文中存储可能不可用。
  }
  writeThemePreferenceCookie(next);
}

export function writeThemePreferenceCookie(next: ThemePreference) {
  if (typeof document === "undefined") {
    return;
  }
  const secure = secureCookieEnabled() ? "; Secure" : "";
  document.cookie = `${THEME_COOKIE_NAME}=${encodeURIComponent(next)}; path=/; max-age=31536000; SameSite=Lax${secure}`;
}

export function syncThemePreferenceFromStorage(): boolean {
  const stored = readStoredThemePreference();
  if (!stored) {
    return false;
  }
  const cookiePreference = readThemePreferenceFromDocumentCookie();
  if (cookiePreference !== stored) {
    writeThemePreferenceCookie(stored);
    return true;
  }
  return false;
}

export { getServerThemeSnapshot };
