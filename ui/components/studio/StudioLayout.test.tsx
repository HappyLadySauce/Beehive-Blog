import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { StudioLayout } from "./StudioLayout";

const pathname = vi.hoisted(() => ({ value: "/studio/content" }));

vi.mock("next/navigation", () => ({
  usePathname: () => pathname.value
}));

describe("StudioLayout", () => {
  it("renders the Studio sidebar on regular Studio routes", () => {
    pathname.value = "/studio/content";

    render(<StudioLayout><div>content child</div></StudioLayout>);

    expect(screen.getByLabelText("Studio 侧边导航")).toBeInTheDocument();
    expect(screen.getByText("content child")).toBeInTheDocument();
  });

  it("renders content editor routes without the Studio sidebar", () => {
    pathname.value = "/studio/content/9/edit";

    render(<StudioLayout><div>editor child</div></StudioLayout>);

    expect(screen.queryByLabelText("Studio 侧边导航")).not.toBeInTheDocument();
    expect(screen.getByText("editor child")).toBeInTheDocument();
  });
});
