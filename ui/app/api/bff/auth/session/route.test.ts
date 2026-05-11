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

import { GET } from "./route";

describe("BFF auth session route", () => {
  beforeEach(() => {
    cookieValues.access = undefined;
    cookieValues.refresh = undefined;
    vi.restoreAllMocks();
  });

  it("verifies the access cookie with Go before returning an authenticated session", async () => {
    cookieValues.access = "access-token";
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ uid: 42, role: "admin", exp: 4_102_444_800, sid: 7 }));
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET();
    const body = await response.json();

    expect(response.status).toBe(200);
    expect(body.data).toMatchObject({ isAuthenticated: true, isAdmin: true, role: "admin" });
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/auth/session",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({ authorization: "Bearer access-token" })
      })
    );
  });

  it("does not trust an unverified access cookie when Go rejects it", async () => {
    cookieValues.access = "forged-admin-token";
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ code: 401, message: "invalid or expired access token", data: null }), {
        status: 401,
        headers: { "content-type": "application/json" }
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET();
    const body = await response.json();

    expect(response.status).toBe(200);
    expect(body.data).toMatchObject({ isAuthenticated: false, isAdmin: false });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});

function jsonResponse(data: unknown) {
  return new Response(JSON.stringify({ code: 200, message: "success", data }), {
    status: 200,
    headers: { "content-type": "application/json" }
  });
}
