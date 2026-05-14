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

import { GET, PATCH } from "./route";

const oauthPayload = {
  revision: 9,
  github_oauth2: {
    enabled: true,
    client_id: "client-id",
    client_secret_set: true,
    redirect_url: "http://localhost:3000/auth/github/callback",
    auth_url: "https://github.com/login/oauth/authorize",
    token_url: "https://github.com/login/oauth/access_token",
    user_info_url: "https://api.github.com/user",
    allow_non_github_endpoints: false
  }
};

describe("BFF settings GitHub OAuth2 route", () => {
  beforeEach(() => {
    cookieValues.access = "access-token";
    cookieValues.refresh = undefined;
    vi.restoreAllMocks();
  });

  it("forwards GET GitHub OAuth2 settings with the HttpOnly access cookie", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(oauthPayload));
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET();
    const body = await response.json();

    expect(response.status).toBe(200);
    expect(body.data).toEqual(oauthPayload);
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/settings/github-oauth2",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({ authorization: "Bearer access-token" })
      })
    );
  });

  it("forwards PATCH GitHub OAuth2 settings body", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(oauthPayload));
    vi.stubGlobal("fetch", fetchMock);

    const response = await PATCH(
      new Request("http://localhost/api/bff/settings/github-oauth2", {
        method: "PATCH",
        body: JSON.stringify({ enabled: true, client_id: "client-id" })
      })
    );

    expect(response.status).toBe(200);
    const forwarded = fetchMock.mock.calls[0][1] as RequestInit;
    expect(JSON.parse(String(forwarded.body))).toEqual({ enabled: true, client_id: "client-id" });
  });

  it("returns 401 when no access or refresh session is present", async () => {
    cookieValues.access = undefined;
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET();
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
