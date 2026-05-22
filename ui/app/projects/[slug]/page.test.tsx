import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/lib/api/content", () => ({
  getPublicContent: vi.fn(async () => ({
    id: 3,
    type: "project",
    typeLabel: "项目",
    href: "/projects/public-project",
    slug: "public-project",
    title: "公开项目",
    description: "项目摘要",
    body: "## 项目目标\n\n- 可验证结果",
    publishedAt: "2026-05-21T00:00:00.000Z",
    tags: ["Project"],
    categories: [{ id: 3, name: "项目", slug: "project" }],
    category: { id: 3, name: "项目", slug: "project" },
    viewCount: 18,
    readingMinutes: 2
  })),
  getPublicReaderContext: vi.fn(async () => ({
    related: [],
    recent: [],
    comments: { items: [], total: 0, page: 1, page_size: 10 }
  })),
  getPublicSiteOverview: vi.fn(async () => ({
    latest: [],
    featured: [],
    recent: [],
    categories: [],
    tags: [],
    archives: [],
    stats: { articles: 0, notes: 0, projects: 1, views: 18, tags: 1 },
    author: { name: "安和鱼", description: "生活明朗，万物可爱" },
    generated_at: "2026-05-22T00:00:00.000Z"
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
