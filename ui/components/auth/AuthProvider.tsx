"use client";

import { createContext, useContext, useEffect, useMemo, useState } from "react";

import type { StoredSession } from "@/lib/auth/session";
import { clearSession, readSession, saveSession } from "@/lib/auth/session";
import type { AuthPayload } from "@/lib/api/types";

type AuthContextValue = {
  session: StoredSession | null;
  isAuthenticated: boolean;
  setAuth: (payload: AuthPayload) => void;
  clearAuth: () => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [session, setSession] = useState<StoredSession | null>(() => readSession());

  useEffect(() => {
    const onChange = () => setSession(readSession());
    window.addEventListener("storage", onChange);
    window.addEventListener("beehive-auth-changed", onChange);
    return () => {
      window.removeEventListener("storage", onChange);
      window.removeEventListener("beehive-auth-changed", onChange);
    };
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      session,
      isAuthenticated: Boolean(session?.token.access_token),
      setAuth(payload) {
        saveSession(payload);
        setSession(readSession());
      },
      clearAuth() {
        clearSession();
        setSession(null);
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
