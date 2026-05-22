import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/lib/api/content", () => ({
  getPublicContent: vi.fn(async () => ({
    id: 2,
    type: "note",
    typeLabel: "笔记",
    href: "/notes/public-note",
    slug: "public-note",
    title: "公开笔记",
    description: "笔记摘要",
    body: "## 笔记标题\n\n- 记录项",
    publishedAt: "2026-05-21T00:00:00.000Z",
    tags: ["Note"],
    categories: [{ id: 2, name: "生活", slug: "life" }],
    category: { id: 2, name: "生活", slug: "life" },
    viewCount: 12,
    readingMinutes: 1
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
    stats: { articles: 0, notes: 1, projects: 0, views: 12, tags: 1 },
    author: { name: "安和鱼", description: "生活明朗，万物可爱" },
    generated_at: "2026-05-22T00:00:00.000Z"
  })),
  listPublicContents: vi.fn(async () => [])
}));

import NoteDetailPage from "./page";

describe("NoteDetailPage", () => {
  it("renders markdown body without exposing raw heading syntax", async () => {
    render(await NoteDetailPage({ params: Promise.resolve({ slug: "public-note" }) }));

    expect(screen.getByRole("heading", { level: 2, name: "笔记标题" })).toBeInTheDocument();
    expect(screen.getByText("记录项")).toBeInTheDocument();
    expect(screen.queryByText("## 笔记标题")).not.toBeInTheDocument();
  });
});
