import { beforeEach, describe, expect, it, vi } from "vitest";

import { createTag, deleteTag, listTags, updateTag } from "./tags";

describe("tags API client", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("serializes list filters", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ items: [], total: 0, page: 1, page_size: 20 }));
    vi.stubGlobal("fetch", fetchMock);

    await listTags({ page: 1, page_size: 20, search: "ai", status: "active" });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/bff/tags?page=1&page_size=20&status=active&search=ai",
      expect.objectContaining({ method: "GET" })
    );
  });

  it("uses expected mutation routes", async () => {
    const fetchMock = vi.fn().mockImplementation(() => Promise.resolve(jsonResponse({ id: 3 })));
    vi.stubGlobal("fetch", fetchMock);

    await createTag({ name: "AI", slug: "ai" });
    await updateTag(3, { color: "#2f8f79" });
    await deleteTag(3);

    expect(fetchMock.mock.calls.map((call) => [call[0], (call[1] as RequestInit).method])).toEqual([
      ["/api/bff/tags", "POST"],
      ["/api/bff/tags/3", "PATCH"],
      ["/api/bff/tags/3", "DELETE"]
    ]);
  });
});

function jsonResponse(data: unknown) {
  return new Response(JSON.stringify({ code: 200, message: "success", data }), {
    status: 200,
    headers: { "content-type": "application/json" }
  });
}
