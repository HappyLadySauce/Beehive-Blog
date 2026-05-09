"use client";

import { createContext, useContext, useMemo, useSyncExternalStore } from "react";

import type { JwtClaims, StoredSession } from "@/lib/auth/session";
import { clearSession, isAdminClaims, isExpiredClaims, readSession, saveSession } from "@/lib/auth/session";
import type { AuthPayload } from "@/lib/api/types";

type AuthContextValue = {
  session: StoredSession | null;
  claims: JwtClaims | null;
  role: string | undefined;
  isAuthenticated: boolean;
  isAdmin: boolean;
  isTokenExpired: boolean;
  setAuth: (payload: AuthPayload) => void;
  clearAuth: () => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

function subscribeAuthStore(onStoreChange: () => void) {
  window.addEventListener("storage", onStoreChange);
  window.addEventListener("beehive-auth-changed", onStoreChange);
  return () => {
    window.removeEventListener("storage", onStoreChange);
    window.removeEventListener("beehive-auth-changed", onStoreChange);
  };
}

function getAuthSnapshot() {
  return readSession();
}

function getServerAuthSnapshot() {
  return null;
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const session = useSyncExternalStore(subscribeAuthStore, getAuthSnapshot, getServerAuthSnapshot);

  const value = useMemo<AuthContextValue>(
    () => ({
      session,
      claims: session?.claims ?? null,
      role: session?.claims?.role,
      isAuthenticated: Boolean(session?.token.access_token && session.claims?.role && !isExpiredClaims(session.claims)),
      isAdmin: isAdminClaims(session?.claims),
      isTokenExpired: isExpiredClaims(session?.claims),
      setAuth(payload) {
        saveSession(payload);
      },
      clearAuth() {
        clearSession();
      }
    }),
    [session]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
}
