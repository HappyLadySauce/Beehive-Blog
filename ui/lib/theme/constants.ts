export const THEME_STORAGE_KEY = "beehive.theme";
export const THEME_COOKIE_NAME = "beehive.theme";

export type ThemePreference = "system" | "light" | "dark";
export type ResolvedTheme = "light" | "dark";

export const DEFAULT_THEME_PREFERENCE: ThemePreference = "system";
export const DEFAULT_SYSTEM_THEME: ResolvedTheme = "light";

export function isThemePreference(value: string | null | undefined): value is ThemePreference {
  return value === "system" || value === "light" || value === "dark";
}
