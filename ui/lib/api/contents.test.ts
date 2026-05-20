import { beforeEach, describe, expect, it, vi } from "vitest";

import { createContent, listContents, setContentTags, transitionContentStatus, updateContent } from "./contents";

describe("contents API client", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("serializes list filters", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ items: [], total: 0, page: 2, page_size: 20 }));
    vi.stubGlobal("fetch", fetchMock);

    await listContents({
      page: 2,
      page_size: 20,
      search: "draft",
      status: "review",
      tag_id: 7,
      type: "article",
      visibility: "private"
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/bff/contents?page=2&page_size=20&type=article&status=review&visibility=private&tag_id=7&search=draft",
      expect.objectContaining({ method: "GET" })
    );
  });

  it("uses expected mutation routes", async () => {
    const fetchMock = vi.fn().mockImplementation(() => Promise.resolve(jsonResponse({ id: 10 })));
    vi.stubGlobal("fetch", fetchMock);

    await createContent({ slug: "hello", title: "Hello", type: "article" });
    await updateContent(10, { title: "Updated" });
    await transitionContentStatus(10, { status: "published" });
    await setContentTags(10, { tag_ids: [1, 2] });

    expect(fetchMock.mock.calls.map((call) => [call[0], (call[1] as RequestInit).method])).toEqual([
      ["/api/bff/contents", "POST"],
      ["/api/bff/contents/10", "PATCH"],
      ["/api/bff/contents/10/status", "PATCH"],
      ["/api/bff/contents/10/tags", "PUT"]
    ]);
  });
});

function jsonResponse(data: unknown) {
  return new Response(JSON.stringify({ code: 200, message: "success", data }), {
    status: 200,
    headers: { "content-type": "application/json" }
  });
}
