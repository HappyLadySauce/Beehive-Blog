import { afterEach, describe, expect, it, vi } from "vitest";

describe("public content API", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
    vi.unstubAllGlobals();
    vi.resetModules();
  });

  it("lists public article posts from the Go content API", async () => {
    vi.stubEnv("BEEHIVE_API_BASE_URL", "http://go.test");
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse({
        items: [
          {
            id: 7,
            type: "article",
            title: "真实文章",
            slug: "real-post",
            excerpt: "来自后端的摘要",
            published_at: "2026-05-20T00:00:00.000Z",
            reading_time_minutes: 6,
            tags: [{ name: "Backend", slug: "backend" }],
            created_at: "2026-05-19T00:00:00.000Z",
            updated_at: "2026-05-20T00:00:00.000Z"
          }
        ],
        total: 1,
        page: 1,
        page_size: 20
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    const { listPublicPosts } = await import("./content");
    const posts = await listPublicPosts();

    expect(fetchMock).toHaveBeenCalledWith(
      "http://go.test/api/v1/contents?page=1&page_size=20&type=article",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({ accept: "application/json" })
      })
    );
    expect(posts).toEqual([
      {
        id: 7,
        slug: "real-post",
        title: "真实文章",
        description: "来自后端的摘要",
        body: "",
        publishedAt: "2026-05-20T00:00:00.000Z",
        tags: ["Backend"],
        categories: [],
        category: undefined,
        coverUrl: undefined,
        authorUsername: undefined,
        viewCount: 0,
        readingMinutes: 6
      }
    ]);
  });

  it("lists public notes and projects with their type-specific routes", async () => {
    vi.stubEnv("BEEHIVE_API_BASE_URL", "http://go.test");
    const fetchMock = vi.fn().mockImplementation(() =>
      Promise.resolve(
        jsonResponse({
          items: [
            {
              id: 11,
              type: "note",
              title: "公开笔记",
              slug: "public-note",
              excerpt: "笔记摘要",
              published_at: "2026-05-21T00:00:00.000Z",
              reading_time_minutes: 1,
              tags: [],
              created_at: "2026-05-21T00:00:00.000Z",
              updated_at: "2026-05-21T00:00:00.000Z"
            }
          ],
          total: 1,
          page: 1,
          page_size: 20
        })
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const { listPublicContents } = await import("./content");
    const notes = await listPublicContents("note");
    const projects = await listPublicContents("project");

    expect(fetchMock.mock.calls.map((call) => call[0])).toEqual([
      "http://go.test/api/v1/contents?page=1&page_size=20&type=note",
      "http://go.test/api/v1/contents?page=1&page_size=20&type=project"
    ]);
    expect(notes[0]).toMatchObject({ type: "note", typeLabel: "笔记", href: "/notes/public-note" });
    expect(projects[0]).toMatchObject({ type: "project", typeLabel: "项目", href: "/projects/public-note" });
  });

  it("loads a public post detail by slug and then by id", async () => {
    vi.stubEnv("BEEHIVE_API_BASE_URL", "http://go.test");
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse({
          items: [
            {
              id: 9,
              type: "article",
              title: "Markdown 文章",
              slug: "markdown-post",
              excerpt: null,
              published_at: "2026-05-21T00:00:00.000Z",
              reading_time_minutes: 2,
              tags: [],
              created_at: "2026-05-21T00:00:00.000Z",
              updated_at: "2026-05-21T00:00:00.000Z"
            }
          ],
          total: 1,
          page: 1,
          page_size: 1
        })
      )
      .mockResolvedValueOnce(
        jsonResponse({
          id: 9,
          type: "article",
          title: "Markdown 文章",
          slug: "markdown-post",
          excerpt: null,
          body: "## 二级标题\n\n正文内容",
          published_at: "2026-05-21T00:00:00.000Z",
          reading_time_minutes: 2,
          tags: [{ name: "Markdown" }],
          created_at: "2026-05-21T00:00:00.000Z",
          updated_at: "2026-05-21T00:00:00.000Z"
        })
      );
    vi.stubGlobal("fetch", fetchMock);

    const { getPublicContent, getPublicPost } = await import("./content");
    const post = await getPublicPost("markdown-post");
    fetchMock.mockClear();
    fetchMock
      .mockResolvedValueOnce(
        jsonResponse({
          items: [
            {
              id: 10,
              type: "project",
              title: "项目",
              slug: "project-post",
              excerpt: "项目摘要",
              published_at: "2026-05-21T00:00:00.000Z",
              reading_time_minutes: 2,
              tags: [],
              created_at: "2026-05-21T00:00:00.000Z",
              updated_at: "2026-05-21T00:00:00.000Z"
            }
          ],
          total: 1,
          page: 1,
          page_size: 1
        })
      )
      .mockResolvedValueOnce(
        jsonResponse({
          id: 10,
          type: "project",
          title: "项目",
          slug: "project-post",
          excerpt: "项目摘要",
          body: "## 项目目标",
          published_at: "2026-05-21T00:00:00.000Z",
          reading_time_minutes: 2,
          tags: [],
          created_at: "2026-05-21T00:00:00.000Z",
          updated_at: "2026-05-21T00:00:00.000Z"
        })
      );
    const project = await getPublicContent("project", "project-post");

    expect(post).toMatchObject({
      slug: "markdown-post",
      title: "Markdown 文章",
      description: "二级标题 正文内容",
      body: "## 二级标题\n\n正文内容",
      tags: ["Markdown"]
    });
    expect(project).toMatchObject({ type: "project", href: "/projects/project-post", body: "## 项目目标" });
  });

  it("uses seeded fallback posts only outside production when the backend fails", async () => {
    vi.stubEnv("NODE_ENV", "development");
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));

    const { listPublicPosts } = await import("./content");

    await expect(listPublicPosts()).resolves.toEqual(
      expect.arrayContaining([expect.objectContaining({ slug: "ai-assisted-writing-loop" })])
    );
  });

  it("does not expose fallback posts in production when the backend fails", async () => {
    vi.stubEnv("NODE_ENV", "production");
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));

    const { listPublicPosts, getPublicPost } = await import("./content");

    await expect(listPublicPosts()).resolves.toEqual([]);
    await expect(getPublicPost("ai-assisted-writing-loop")).resolves.toBeNull();
  });

  it("uses seeded fallback post details outside production when the backend returns an empty slug result", async () => {
    vi.stubEnv("NODE_ENV", "development");
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(jsonResponse({ items: [], total: 0, page: 1, page_size: 1 }))
    );

    const { getPublicPost } = await import("./content");

    await expect(getPublicPost("ai-assisted-writing-loop")).resolves.toMatchObject({
      slug: "ai-assisted-writing-loop",
      title: "AI 协作写作回路"
    });
  });

  it("returns null when a public slug is not found", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(jsonResponse({ items: [], total: 0, page: 1, page_size: 1 }))
    );

    const { getPublicPost } = await import("./content");

    await expect(getPublicPost("missing")).resolves.toBeNull();
  });
});

function jsonResponse(data: unknown) {
  return new Response(JSON.stringify({ code: 200, message: "success", data }), {
    status: 200,
    headers: { "content-type": "application/json" }
  });
}
