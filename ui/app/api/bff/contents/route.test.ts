import { beforeEach, describe, expect, it, vi } from "vitest";

const cookieValues = vi.hoisted(() => ({
  access: "access-token",
  refresh: undefined as string | undefined
}));

vi.mock("next/headers", () => ({
  cookies: vi.fn(async () => ({
    get(name: string) {
      if (name === "beehive.access" && cookieValues.access) return { value: cookieValues.access };
      if (name === "beehive.refresh" && cookieValues.refresh) return { value: cookieValues.refresh };
      return undefined;
    }
  }))
}));

import { DELETE, PATCH } from "./[id]/route";
import { PATCH as PATCH_STATUS } from "./[id]/status/route";
import { PUT as PUT_TAGS } from "./[id]/tags/route";
import { GET, POST } from "./route";

describe("BFF contents route", () => {
  beforeEach(() => {
    cookieValues.access = "access-token";
    cookieValues.refresh = undefined;
    vi.restoreAllMocks();
  });

  it("forwards list query parameters with the access cookie", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ items: [], total: 0, page: 1, page_size: 20 }));
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(new Request("http://localhost/api/bff/contents?status=draft&page=1"));
    const body = await response.json();

    expect(response.status).toBe(200);
    expect(body.data.items).toEqual([]);
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/contents?status=draft&page=1",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({ authorization: "Bearer access-token" })
      })
    );
  });

  it("forwards create, update, status, tag, and delete mutations", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ id: 9 }));
    vi.stubGlobal("fetch", fetchMock);
    const context = { params: Promise.resolve({ id: "9" }) };

    await POST(new Request("http://localhost/api/bff/contents", { method: "POST", body: JSON.stringify({ title: "T" }) }));
    await PATCH(new Request("http://localhost/api/bff/contents/9", { method: "PATCH", body: JSON.stringify({ title: "U" }) }), context);
    await PATCH_STATUS(
      new Request("http://localhost/api/bff/contents/9/status", { method: "PATCH", body: JSON.stringify({ status: "published" }) }),
      context
    );
    await PUT_TAGS(
      new Request("http://localhost/api/bff/contents/9/tags", { method: "PUT", body: JSON.stringify({ tag_ids: [1] }) }),
      context
    );
    await DELETE(new Request("http://localhost/api/bff/contents/9", { method: "DELETE" }), context);

    expect(fetchMock.mock.calls.map((call) => [call[0], (call[1] as RequestInit).method])).toEqual([
      ["http://localhost:8080/api/v1/contents", "POST"],
      ["http://localhost:8080/api/v1/contents/9", "PATCH"],
      ["http://localhost:8080/api/v1/contents/9/status", "PATCH"],
      ["http://localhost:8080/api/v1/contents/9/tags", "PUT"],
      ["http://localhost:8080/api/v1/contents/9", "DELETE"]
    ]);
  });

  it("maps upstream errors with the original envelope status", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ code: 409, message: "slug conflict", data: null }), {
        status: 409,
        headers: { "content-type": "application/json" }
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/bff/contents", { method: "POST", body: "{}" }));
    const body = await response.json();

    expect(response.status).toBe(409);
    expect(body.message).toBe("slug conflict");
  });
});

function jsonResponse(data: unknown) {
  return new Response(JSON.stringify({ code: 200, message: "success", data }), {
    status: 200,
    headers: { "content-type": "application/json" }
  });
}
