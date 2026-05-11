import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { SiteHeader } from "./SiteHeader";

const authState = vi.hoisted(() => ({
  isAuthenticated: false,
  isAdmin: false,
  loading: false,
  role: undefined as string | undefined,
  claims: null as { uid?: number; role?: string } | null,
  clearAuth: vi.fn()
}));

vi.mock("@/components/auth/AuthProvider", () => ({
  useAuth: () => authState
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    replace: vi.fn()
  })
}));

describe("SiteHeader", () => {
  beforeEach(() => {
    authState.isAuthenticated = false;
    authState.isAdmin = false;
    authState.loading = false;
    authState.role = undefined;
    authState.claims = null;
    authState.clearAuth.mockClear();
  });

  it("shows login for anonymous visitors", () => {
    render(<SiteHeader />);

    expect(screen.getByRole("link", { name: "登录" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Studio" })).not.toBeInTheDocument();
  });

  it("hides Studio from regular users", () => {
    authState.isAuthenticated = true;
    authState.isAdmin = false;
    authState.role = "member";
    authState.claims = { uid: 7, role: "member" };

    render(<SiteHeader />);
    fireEvent.click(screen.getByRole("button", { name: "打开用户菜单" }));

    expect(screen.queryByRole("link", { name: "Studio" })).not.toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "登出" })).toBeInTheDocument();
  });

  it("shows Studio in the user menu for admins", () => {
    authState.isAuthenticated = true;
    authState.isAdmin = true;
    authState.role = "admin";
    authState.claims = { uid: 1, role: "admin" };

    render(<SiteHeader />);
    fireEvent.click(screen.getByRole("button", { name: "打开用户菜单" }));

    expect(screen.getByRole("link", { name: "Studio" })).toHaveAttribute("href", "/studio");
    expect(screen.getByRole("menuitem", { name: "登出" })).toBeInTheDocument();
  });

  it("closes the user menu when clicking outside", () => {
    authState.isAuthenticated = true;
    authState.isAdmin = true;
    authState.role = "admin";
    authState.claims = { uid: 1, role: "admin" };

    render(
      <div>
        <SiteHeader />
        <main data-testid="outside">Outside</main>
      </div>
    );
    fireEvent.click(screen.getByRole("button", { name: "打开用户菜单" }));
    expect(screen.getByRole("menuitem", { name: "登出" })).toBeInTheDocument();

    fireEvent.pointerDown(screen.getByTestId("outside"));

    expect(screen.queryByRole("menuitem", { name: "登出" })).not.toBeInTheDocument();
  });
});
