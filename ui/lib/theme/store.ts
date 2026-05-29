"use client";

import { getClientThemeSnapshot, syncThemePreferenceFromStorage, writeThemePreference } from "./client";
import { SERVER_THEME_SNAPSHOT, type ThemeSnapshot } from "./resolve";
import { THEME_STORAGE_KEY, type ThemePreference } from "./constants";

function snapshotKey(snapshot: ThemeSnapshot) {
  return `${snapshot.preference}:${snapshot.systemTheme}`;
}

let cachedClientSnapshot: ThemeSnapshot = SERVER_THEME_SNAPSHOT;
let cachedClientSnapshotKey = snapshotKey(SERVER_THEME_SNAPSHOT);

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
  const fresh = getClientThemeSnapshot();
  const key = snapshotKey(fresh);
  if (key === cachedClientSnapshotKey) {
    return cachedClientSnapshot;
  }
  cachedClientSnapshotKey = key;
  cachedClientSnapshot = fresh;
  return cachedClientSnapshot;
}

export function getThemeServerSnapshot(): ThemeSnapshot {
  return SERVER_THEME_SNAPSHOT;
}

export function setThemePreference(next: ThemePreference) {
  writeThemePreference(next);
  cachedClientSnapshotKey = "";
  emitChange();
}

export function ensureThemePreferenceSynced() {
  if (syncThemePreferenceFromStorage()) {
    cachedClientSnapshotKey = "";
    emitChange();
  }
}
