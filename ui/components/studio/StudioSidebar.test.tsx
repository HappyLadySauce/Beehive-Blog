import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { StudioSidebar } from "./StudioSidebar";

const pathname = vi.hoisted(() => ({ value: "/studio/content" }));

vi.mock("next/navigation", () => ({
  usePathname: () => pathname.value
}));

describe("StudioSidebar", () => {
  it("marks the active Studio route", () => {
    render(<StudioSidebar />);

    expect(screen.getByRole("link", { name: "内容" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("link", { name: "总览" })).not.toHaveAttribute("aria-current");
  });

  it("supports mobile navigation toggling", () => {
    render(<StudioSidebar />);

    const toggle = screen.getByRole("button", { name: "导航" });
    expect(toggle).toHaveAttribute("aria-expanded", "false");
    fireEvent.click(toggle);
    expect(toggle).toHaveAttribute("aria-expanded", "true");
  });
});
