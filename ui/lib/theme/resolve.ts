import {
  DEFAULT_SYSTEM_THEME,
  DEFAULT_THEME_PREFERENCE,
  isThemePreference,
  type ResolvedTheme,
  type ThemePreference,
  THEME_COOKIE_NAME
} from "./constants";

export type ThemeSnapshot = {
  preference: ThemePreference;
  systemTheme: ResolvedTheme;
};

export function resolveThemeFromPreference(
  preference: ThemePreference,
  systemTheme: ResolvedTheme = DEFAULT_SYSTEM_THEME
): ResolvedTheme {
  return preference === "system" ? systemTheme : preference;
}

function systemThemeFromClientHint(value: string | null | undefined): ResolvedTheme | null {
  if (value === "dark") return "dark";
  if (value === "light") return "light";
  return null;
}

export function readThemePreferenceFromCookieValue(value: string | undefined | null): ThemePreference {
  return isThemePreference(value) ? value : DEFAULT_THEME_PREFERENCE;
}

type ThemeCookieReader = {
  get(name: string): { value: string } | undefined;
};

export function resolveThemeFromCookies(cookieStore: ThemeCookieReader, headers?: Pick<Headers, "get">): ResolvedTheme {
  const preference = readThemePreferenceFromCookieValue(cookieStore.get(THEME_COOKIE_NAME)?.value);
  if (preference === "light" || preference === "dark") {
    return preference;
  }
  const hinted = systemThemeFromClientHint(headers?.get("sec-ch-prefers-color-scheme"));
  return hinted ?? DEFAULT_SYSTEM_THEME;
}

export function getServerThemeSnapshot(): ThemeSnapshot {
  return { preference: DEFAULT_THEME_PREFERENCE, systemTheme: DEFAULT_SYSTEM_THEME };
}
