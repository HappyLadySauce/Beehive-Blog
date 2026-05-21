import { describe, expect, it, vi } from "vitest";

vi.mock("@/lib/api/content", () => ({
  listPublicPosts: vi.fn(async () => [
    {
      slug: "article-one",
      title: "文章",
      description: "文章摘要",
      body: "",
      publishedAt: "2026-05-20T00:00:00.000Z",
      tags: [],
      readingMinutes: 1
    }
  ]),
  listPublicContents: vi.fn(async (type: string) => [
    {
      type,
      typeLabel: type === "note" ? "笔记" : "项目",
      href: type === "note" ? "/notes/note-one" : "/projects/project-one",
      slug: type === "note" ? "note-one" : "project-one",
      title: type,
      description: type,
      body: "",
      publishedAt: "2026-05-21T00:00:00.000Z",
      tags: [],
      readingMinutes: 1
    }
  ])
}));

import sitemap from "./sitemap";

describe("sitemap", () => {
  it("includes article, note, and project public routes", async () => {
    const urls = (await sitemap()).map((entry) => entry.url);

    expect(urls).toEqual(
      expect.arrayContaining([
        "http://localhost:3000/posts",
        "http://localhost:3000/notes",
        "http://localhost:3000/projects",
        "http://localhost:3000/posts/article-one",
        "http://localhost:3000/notes/note-one",
        "http://localhost:3000/projects/project-one"
      ])
    );
  });
});
