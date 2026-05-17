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

import { GET, POST } from "./route";

describe("BFF attachments route", () => {
  beforeEach(() => {
    cookieValues.access = "access-token";
    cookieValues.refresh = undefined;
    vi.restoreAllMocks();
  });

  it("forwards list query parameters with the access cookie", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ items: [] }));
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET(new Request("http://localhost/api/bff/attachments?purpose=content&limit=20"));
    const body = await response.json();

    expect(response.status).toBe(200);
    expect(body.data).toEqual({ items: [] });
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/attachments?purpose=content&limit=20",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({ authorization: "Bearer access-token" })
      })
    );
  });

  it("forwards multipart uploads without forcing a JSON content type", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ id: 1 }));
    vi.stubGlobal("fetch", fetchMock);

    const formData = new FormData();
    formData.set("file", new File(["hello"], "hello.txt", { type: "text/plain" }));
    const response = await POST(
      new Request("http://localhost/api/bff/attachments", {
        method: "POST",
        body: formData
      })
    );

    expect(response.status).toBe(200);
    const forwarded = fetchMock.mock.calls[0][1] as RequestInit;
    expect(String(forwarded.body)).toBe("[object FormData]");
    expect(forwarded.headers).not.toMatchObject({ "content-type": "application/json" });
  });
});

function jsonResponse(data: unknown) {
  return new Response(JSON.stringify({ code: 200, message: "success", data }), {
    status: 200,
    headers: { "content-type": "application/json" }
  });
}
