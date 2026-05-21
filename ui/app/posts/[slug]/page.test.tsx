import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/lib/api/content", () => ({
  getPublicPost: vi.fn(async () => ({
    slug: "markdown-post",
    title: "Markdown 文章",
    description: "Markdown 摘要",
    body: "## 二级标题\n\n- 第一项\n- 第二项",
    publishedAt: "2026-05-21T00:00:00.000Z",
    tags: ["Markdown"],
    readingMinutes: 2
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
