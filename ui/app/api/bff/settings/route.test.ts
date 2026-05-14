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

const settingsPayload = {
  revision: 3,
  email: {
    enabled: true,
    host: "smtp.example.com",
    port: 587,
    username: "robot",
    password_set: true,
    from: "robot@example.com",
    from_name: "Beehive",
    tls: "starttls"
  }
};

describe("BFF settings route", () => {
  beforeEach(() => {
    cookieValues.access = "access-token";
    cookieValues.refresh = undefined;
    vi.restoreAllMocks();
  });

  it("forwards GET settings with the HttpOnly access cookie", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(settingsPayload));
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET();
    const body = await response.json();

    expect(response.status).toBe(200);
    expect(body.data).toEqual(settingsPayload);
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/settings/email",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({ authorization: "Bearer access-token" })
      })
    );
  });

  it("forwards PATCH settings body without adding a password field", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(settingsPayload));
    vi.stubGlobal("fetch", fetchMock);

    const response = await PATCH(
      new Request("http://localhost/api/bff/settings", {
        method: "PATCH",
        body: JSON.stringify({ email: { enabled: true, host: "smtp.example.com" } })
      })
    );

    expect(response.status).toBe(200);
    const forwarded = fetchMock.mock.calls[0][1] as RequestInit;
    expect(JSON.parse(String(forwarded.body))).toEqual({ email: { enabled: true, host: "smtp.example.com" } });
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

  it("maps upstream forbidden responses without retrying refresh", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ code: 403, message: "forbidden", data: null }), {
        status: 403,
        headers: { "content-type": "application/json" }
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET();
    const body = await response.json();

    expect(response.status).toBe(403);
    expect(body.message).toBe("forbidden");
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});

function jsonResponse(data: unknown) {
  return new Response(JSON.stringify({ code: 200, message: "success", data }), {
    status: 200,
    headers: { "content-type": "application/json" }
  });
}
