import { beforeEach, describe, expect, it, vi } from "vitest";

const cookieValues = vi.hoisted(() => ({
  access: undefined as string | undefined,
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

import { POST } from "./route";

describe("BFF settings email test route", () => {
  beforeEach(() => {
    cookieValues.access = "access-token";
    cookieValues.refresh = undefined;
    vi.restoreAllMocks();
  });

  it("forwards SMTP test requests with the HttpOnly access cookie", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ recipient: "reader@example.com" }));
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(
      new Request("http://localhost/api/bff/settings/email/test", {
        method: "POST",
        body: JSON.stringify({ recipient: "reader@example.com" })
      })
    );
    const body = await response.json();

    expect(response.status).toBe(200);
    expect(body.data).toEqual({ recipient: "reader@example.com" });
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/settings/email/test",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({ authorization: "Bearer access-token" }),
        body: JSON.stringify({ recipient: "reader@example.com" })
      })
    );
  });

  it("returns 401 when no access or refresh session is present", async () => {
    cookieValues.access = undefined;
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(
      new Request("http://localhost/api/bff/settings/email/test", {
        method: "POST",
        body: JSON.stringify({ recipient: "reader@example.com" })
      })
    );
    const body = await response.json();

    expect(response.status).toBe(401);
    expect(body.message).toBe("Missing authenticated session");
    expect(fetchMock).not.toHaveBeenCalled();
  });
});

function jsonResponse(data: unknown) {
  return new Response(JSON.stringify({ code: 200, message: "success", data }), {
    status: 200,
    headers: { "content-type": "application/json" }
  });
}
