import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { StudioShell } from "./StudioShell";

const replace = vi.hoisted(() => vi.fn());
const authState = vi.hoisted(() => ({
  isAuthenticated: false,
  isAdmin: false,
  session: null
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace })
}));

vi.mock("@/components/auth/AuthProvider", () => ({
  useAuth: () => authState
}));

describe("StudioShell", () => {
  beforeEach(() => {
    replace.mockClear();
    authState.isAuthenticated = false;
    authState.isAdmin = false;
    authState.session = null;
  });

  it("redirects anonymous visitors to login", async () => {
    render(<StudioShell />);

    await waitFor(() => expect(replace).toHaveBeenCalledWith("/login?next=/studio"));
    expect(screen.queryByRole("heading", { name: "内容管理与发布闸门" })).not.toBeInTheDocument();
  });

  it("redirects regular users away from Studio", async () => {
    authState.isAuthenticated = true;
    authState.isAdmin = false;

    render(<StudioShell />);

    await waitFor(() => expect(replace).toHaveBeenCalledWith("/"));
    expect(screen.queryByRole("heading", { name: "内容管理与发布闸门" })).not.toBeInTheDocument();
  });

  it("renders the Studio workspace for admins", () => {
    authState.isAuthenticated = true;
    authState.isAdmin = true;

    render(<StudioShell />);

    expect(screen.getByRole("heading", { name: "内容管理与发布闸门" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "登出" })).not.toBeInTheDocument();
  });
});
