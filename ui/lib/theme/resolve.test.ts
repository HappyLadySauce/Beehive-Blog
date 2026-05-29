import { describe, expect, it } from "vitest";

import { THEME_COOKIE_NAME } from "./constants";
import { resolveThemeFromCookies, resolveThemeFromPreference, readThemePreferenceFromCookieValue } from "./resolve";

describe("resolveThemeFromPreference", () => {
  it("returns explicit light and dark preferences", () => {
    expect(resolveThemeFromPreference("light")).toBe("light");
    expect(resolveThemeFromPreference("dark")).toBe("dark");
  });

  it("uses system theme when preference is system", () => {
    expect(resolveThemeFromPreference("system", "dark")).toBe("dark");
    expect(resolveThemeFromPreference("system", "light")).toBe("light");
  });
});

describe("resolveThemeFromCookies", () => {
  it("reads explicit cookie preferences", () => {
    const cookieStore = {
      get: (name: string) => (name === THEME_COOKIE_NAME ? { value: "dark" } : undefined)
    };
    expect(resolveThemeFromCookies(cookieStore)).toBe("dark");
  });

  it("uses client hint for system preference", () => {
    const cookieStore = {
      get: (name: string) => (name === THEME_COOKIE_NAME ? { value: "system" } : undefined)
    };
    const headers = new Headers({ "sec-ch-prefers-color-scheme": "dark" });
    expect(resolveThemeFromCookies(cookieStore, headers)).toBe("dark");
  });

  it("falls back to light when system preference has no hint", () => {
    const cookieStore = {
      get: () => undefined
    };
    expect(resolveThemeFromCookies(cookieStore)).toBe("light");
  });
});

describe("readThemePreferenceFromCookieValue", () => {
  it("defaults invalid values to system", () => {
    expect(readThemePreferenceFromCookieValue("invalid")).toBe("system");
    expect(readThemePreferenceFromCookieValue(undefined)).toBe("system");
  });
});
