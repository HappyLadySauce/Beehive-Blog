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
        slug: "real-post",
        title: "真实文章",
        description: "来自后端的摘要",
        body: "",
        publishedAt: "2026-05-20T00:00:00.000Z",
        tags: ["Backend"],
        readingMinutes: 6
      }
    ]);
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

    const { getPublicPost } = await import("./content");
    const post = await getPublicPost("markdown-post");

    expect(fetchMock.mock.calls.map((call) => call[0])).toEqual([
      "http://go.test/api/v1/contents?page=1&page_size=1&type=article&slug=markdown-post",
      "http://go.test/api/v1/contents/9"
    ]);
    expect(post).toMatchObject({
      slug: "markdown-post",
      title: "Markdown 文章",
      description: "二级标题 正文内容",
      body: "## 二级标题\n\n正文内容",
      tags: ["Markdown"]
    });
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
