import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/lib/api/content", () => ({
  getPublicPost: vi.fn(async () => ({
    id: 1,
    slug: "markdown-post",
    title: "Markdown 文章",
    description: "Markdown 摘要",
    body: "## 二级标题\n\n- 第一项\n- 第二项",
    publishedAt: "2026-05-21T00:00:00.000Z",
    tags: ["Markdown"],
    categories: [{ id: 1, name: "前端", slug: "frontend" }],
    category: { id: 1, name: "前端", slug: "frontend" },
    viewCount: 88,
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
    stats: { articles: 1, notes: 0, projects: 0, views: 88, tags: 1 },
    author: { name: "安和鱼", description: "生活明朗，万物可爱" },
    generated_at: "2026-05-22T00:00:00.000Z"
  })),
  listPublicPosts: vi.fn(async () => [])
}));

import PostDetailPage from "./page";

describe("PostDetailPage", () => {
  it("renders markdown body as public HTML instead of raw markdown source", async () => {
    render(await PostDetailPage({ params: Promise.resolve({ slug: "markdown-post" }) }));

    expect(screen.getByRole("heading", { level: 2, name: "二级标题" })).toBeInTheDocument();
    expect(screen.getByText("第一项")).toBeInTheDocument();
    expect(screen.queryByText("## 二级标题")).not.toBeInTheDocument();
  });
});
