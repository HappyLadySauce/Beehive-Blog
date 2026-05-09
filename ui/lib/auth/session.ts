import type { AuthPayload, AuthToken } from "@/lib/api/types";

const storageKey = "beehive.auth";
let cachedRawSession: string | null | undefined;
let cachedSession: StoredSession | null = null;

export type JwtClaims = {
  uid?: number;
  role?: string;
  exp?: number;
  sid?: number;
  use?: string;
};

export type StoredSession = {
  token: AuthToken;
  claims?: JwtClaims;
  issuedAt: number;
};

export function saveSession(payload: AuthPayload) {
  const claims = decodeJwtClaims(payload.token.access_token);
  const session: StoredSession = {
    token: payload.token,
    claims: claims ?? undefined,
    issuedAt: Date.now()
  };
  window.localStorage.setItem(storageKey, JSON.stringify(session));
  window.dispatchEvent(new Event("beehive-auth-changed"));
}

export function decodeJwtClaims(token: string): JwtClaims | null {
  const payload = token.split(".")[1];
  if (!payload) return null;

  try {
    const normalized = payload.replace(/-/g, "+").replace(/_/g, "/");
    const padded = normalized.padEnd(normalized.length + ((4 - (normalized.length % 4)) % 4), "=");
    const json =
      typeof atob === "function"
        ? decodeURIComponent(
            Array.from(atob(padded), (char) => `%${char.charCodeAt(0).toString(16).padStart(2, "0")}`).join("")
          )
        : Buffer.from(padded, "base64").toString("utf8");
    const claims = JSON.parse(json) as JwtClaims;
    return typeof claims === "object" && claims !== null ? claims : null;
  } catch {
    return null;
  }
}

export function isExpiredClaims(claims: JwtClaims | null | undefined) {
  if (!claims?.exp) return true;
  return claims.exp * 1000 <= Date.now();
}

export function isAdminClaims(claims: JwtClaims | null | undefined) {
  return claims?.role === "admin" && !isExpiredClaims(claims);
}

export function readSession(): StoredSession | null {
  if (typeof window === "undefined") return null;

  const raw = window.localStorage.getItem(storageKey);
  if (raw === cachedRawSession) return cachedSession;

  cachedRawSession = raw;
  cachedSession = null;
  if (!raw) return null;

  try {
    const parsed = JSON.parse(raw) as StoredSession;
    if (!parsed.token?.access_token) return null;
    const claims = decodeJwtClaims(parsed.token.access_token);
    if (!claims?.role || isExpiredClaims(claims)) return null;
    cachedSession = { ...parsed, claims };
    return cachedSession;
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
