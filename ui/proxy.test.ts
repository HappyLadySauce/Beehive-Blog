import { NextRequest } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { proxy } from "./proxy";

describe("Studio proxy", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("does not allow a forged admin access cookie when Go rejects it", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ code: 401, message: "invalid or expired access token", data: null }), {
        status: 401,
        headers: { "content-type": "application/json" }
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    const response = await proxy(requestWithCookies("beehive.access=forged-admin-token"));

    expect(response.status).toBeGreaterThanOrEqual(300);
    expect(response.headers.get("location")).toBe("http://localhost/login?next=%2Fstudio%2Fsettings");
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/auth/session",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({ authorization: "Bearer forged-admin-token" })
      })
    );
  });

  it("allows Studio when Go verifies an admin access cookie", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ uid: 1, role: "admin", exp: 4_102_444_800, sid: 9 }));
    vi.stubGlobal("fetch", fetchMock);

    const response = await proxy(requestWithCookies("beehive.access=access-token"));

    expect(response.headers.get("x-middleware-next")).toBe("1");
  });
});

function requestWithCookies(cookie: string) {
  return new NextRequest("http://localhost/studio/settings", {
    headers: {
      cookie
    }
  });
}

function jsonResponse(data: unknown) {
  return new Response(JSON.stringify({ code: 200, message: "success", data }), {
    status: 200,
    headers: { "content-type": "application/json" }
  });
}
