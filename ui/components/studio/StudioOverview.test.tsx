import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { StudioOverview } from "./StudioOverview";

describe("StudioOverview", () => {
  it("renders dashboard metrics and disabled content creation", () => {
    render(<StudioOverview />);

    expect(screen.getByRole("heading", { name: "内容管理与发布闸门" })).toBeInTheDocument();
    expect(screen.getByText("公开内容")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "新建内容" })).toBeDisabled();
    expect(screen.getByText("BFF Cookie 会话已接管")).toBeInTheDocument();
  });
});
