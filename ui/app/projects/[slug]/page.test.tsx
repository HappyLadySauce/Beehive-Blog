import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/lib/api/content", () => ({
  getPublicContent: vi.fn(async () => ({
    type: "project",
    typeLabel: "项目",
    href: "/projects/public-project",
    slug: "public-project",
    title: "公开项目",
    description: "项目摘要",
    body: "## 项目目标\n\n- 可验证结果",
    publishedAt: "2026-05-21T00:00:00.000Z",
    tags: ["Project"],
    readingMinutes: 2
  })),
  listPublicContents: vi.fn(async () => [])
}));

import ProjectDetailPage from "./page";

describe("ProjectDetailPage", () => {
  it("renders markdown body without exposing raw heading syntax", async () => {
    render(await ProjectDetailPage({ params: Promise.resolve({ slug: "public-project" }) }));

    expect(screen.getByRole("heading", { level: 2, name: "项目目标" })).toBeInTheDocument();
    expect(screen.getByText("可验证结果")).toBeInTheDocument();
    expect(screen.queryByText("## 项目目标")).not.toBeInTheDocument();
  });
});
