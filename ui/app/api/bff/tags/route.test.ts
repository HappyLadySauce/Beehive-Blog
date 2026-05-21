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
import { GET, POST } from "./route";

describe("BFF tags route", () => {
  beforeEach(() => {
    cookieValues.access = "access-token";
    cookieValues.refresh = undefined;
    vi.restoreAllMocks();
  });

  it("forwards list query parameters with the access cookie", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ items: [], total: 0, page: 1, page_size: 20 }));
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(new Request("http://localhost/api/bff/tags?search=ai"));
    const body = await response.json();

    expect(response.status).toBe(200);
    expect(body.data.items).toEqual([]);
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/tags?search=ai",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({ authorization: "Bearer access-token" })
      })
    );
  });

  it("forwards create, update, and delete mutations", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ id: 7 }));
    vi.stubGlobal("fetch", fetchMock);
    const context = { params: Promise.resolve({ id: "7" }) };

    await POST(new Request("http://localhost/api/bff/tags", { method: "POST", body: JSON.stringify({ name: "AI" }) }));
    await PATCH(new Request("http://localhost/api/bff/tags/7", { method: "PATCH", body: JSON.stringify({ color: "#2f8f79" }) }), context);
    await DELETE(new Request("http://localhost/api/bff/tags/7", { method: "DELETE" }), context);

    expect(fetchMock.mock.calls.map((call) => [call[0], (call[1] as RequestInit).method])).toEqual([
      ["http://localhost:8080/api/v1/tags", "POST"],
      ["http://localhost:8080/api/v1/tags/7", "PATCH"],
      ["http://localhost:8080/api/v1/tags/7", "DELETE"]
    ]);
  });

  it("keeps upstream conflict messages", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ code: 409, message: "tag is referenced", data: null }), {
        status: 409,
        headers: { "content-type": "application/json" }
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    const response = await DELETE(new Request("http://localhost/api/bff/tags/7", { method: "DELETE" }), {
      params: Promise.resolve({ id: "7" })
    });
    const body = await response.json();

    expect(response.status).toBe(409);
    expect(body.message).toBe("tag is referenced");
  });
});

function jsonResponse(data: unknown) {
  return new Response(JSON.stringify({ code: 200, message: "success", data }), {
    status: 200,
    headers: { "content-type": "application/json" }
  });
}
