import type { AuthPayload, AuthToken } from "@/lib/api/types";

const storageKey = "beehive.auth";

export type StoredSession = {
  token: AuthToken;
  issuedAt: number;
};

export function saveSession(payload: AuthPayload) {
  const session: StoredSession = {
    token: payload.token,
    issuedAt: Date.now()
  };
  window.localStorage.setItem(storageKey, JSON.stringify(session));
  window.dispatchEvent(new Event("beehive-auth-changed"));
}

export function readSession(): StoredSession | null {
  if (typeof window === "undefined") return null;

  const raw = window.localStorage.getItem(storageKey);
  if (!raw) return null;

  try {
    const parsed = JSON.parse(raw) as StoredSession;
    if (!parsed.token?.access_token) return null;
    return parsed;
  } catch {
    clearSession();
    return null;
  }
}

export function clearSession() {
  if (typeof window === "undefined") return;
  window.localStorage.removeItem(storageKey);
  window.dispatchEvent(new Event("beehive-auth-changed"));
}
