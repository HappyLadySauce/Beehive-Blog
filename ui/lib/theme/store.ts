"use client";

import {
  getClientThemeSnapshot,
  getServerThemeSnapshot,
  syncThemePreferenceFromStorage,
  writeThemePreference
} from "./client";
import type { ThemeSnapshot } from "./resolve";
import { THEME_STORAGE_KEY, type ThemePreference } from "./constants";

type Listener = () => void;

const listeners = new Set<Listener>();

function emitChange() {
  listeners.forEach((listener) => listener());
}

export function subscribeThemeStore(listener: Listener) {
  listeners.add(listener);
  if (typeof window === "undefined") {
    return () => listeners.delete(listener);
  }

  const media = window.matchMedia("(prefers-color-scheme: dark)");
  const onMediaChange = () => emitChange();
  const onStorage = (event: StorageEvent) => {
    if (event.key === null || event.key === THEME_STORAGE_KEY) {
      emitChange();
    }
  };

  media.addEventListener("change", onMediaChange);
  window.addEventListener("storage", onStorage);

  return () => {
    listeners.delete(listener);
    media.removeEventListener("change", onMediaChange);
    window.removeEventListener("storage", onStorage);
  };
}

export function getThemeSnapshot(): ThemeSnapshot {
  return getClientThemeSnapshot();
}

export function getThemeServerSnapshot(): ThemeSnapshot {
  return getServerThemeSnapshot();
}

export function setThemePreference(next: ThemePreference) {
  writeThemePreference(next);
  emitChange();
}

export function ensureThemePreferenceSynced() {
  syncThemePreferenceFromStorage();
  emitChange();
}
