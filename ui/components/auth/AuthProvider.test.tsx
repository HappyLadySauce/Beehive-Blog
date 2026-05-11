import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";

import { AuthProvider, useAuth } from "./AuthProvider";
import { sessionFromClaims } from "@/lib/auth/session";

const getSession = vi.hoisted(() => vi.fn());
const logout = vi.hoisted(() => vi.fn());

vi.mock("@/lib/api/auth", () => ({
  getSession,
  logout
}));

function AuthProbe() {
  const auth = useAuth();
  return (
    <div>
      <span>{auth.loading ? "loading" : "ready"}</span>
      <span>{auth.isAuthenticated ? "authenticated" : "anonymous"}</span>
      <span>{auth.isAdmin ? "admin" : "not-admin"}</span>
    </div>
  );
}

describe("AuthProvider", () => {
  beforeEach(() => {
    getSession.mockReset();
    logout.mockReset();
  });

  it("loads an admin session from the BFF", async () => {
    getSession.mockResolvedValue(sessionFromClaims({ uid: 1, role: "admin", exp: 4_102_444_800 }));

    render(
      <AuthProvider>
        <AuthProbe />
      </AuthProvider>
    );

    expect(screen.getByText("loading")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText("ready")).toBeInTheDocument());
    expect(screen.getByText("authenticated")).toBeInTheDocument();
    expect(screen.getByText("admin")).toBeInTheDocument();
  });

  it("falls back to anonymous when session loading fails", async () => {
    getSession.mockRejectedValue(new Error("network failed"));
    vi.spyOn(console, "error").mockImplementation(() => undefined);

    render(
      <AuthProvider>
        <AuthProbe />
      </AuthProvider>
    );

    await waitFor(() => expect(screen.getByText("ready")).toBeInTheDocument());
    expect(screen.getByText("anonymous")).toBeInTheDocument();
  });
});
