"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";

import { getSession, logout } from "@/lib/api/auth";
import type { ClientSession, JwtClaims } from "@/lib/auth/session";
import { sessionFromClaims } from "@/lib/auth/session";

type AuthContextValue = ClientSession & {
  loading: boolean;
  refreshSession: () => Promise<ClientSession>;
  setAuth: (session: ClientSession) => void;
  clearAuth: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);
const anonymousSession = sessionFromClaims(null);
let initialSessionRequest: Promise<ClientSession> | null = null;

function loadInitialSession() {
  if (!initialSessionRequest) {
    initialSessionRequest = getSession().finally(() => {
      initialSessionRequest = null;
    });
  }
  return initialSessionRequest;
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [session, setSession] = useState<ClientSession>(anonymousSession);
  const [loading, setLoading] = useState(true);

  const refreshSession = useCallback(async () => {
    const nextSession = await getSession();
    setSession(nextSession);
    return nextSession;
  }, []);

  useEffect(() => {
    let active = true;

    loadInitialSession()
      .then((nextSession) => {
        if (active) {
          setSession(nextSession);
        }
      })
      .catch((error) => {
        console.error("Failed to load auth session", error);
        if (active) {
          setSession(anonymousSession);
        }
      })
      .finally(() => {
        if (active) {
          setLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      ...session,
      loading,
      refreshSession,
      setAuth(nextSession) {
        setSession(nextSession);
      },
      async clearAuth() {
        try {
          await logout();
        } finally {
          setSession(anonymousSession);
        }
      }
    }),
    [loading, refreshSession, session]
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

export type { ClientSession, JwtClaims };
