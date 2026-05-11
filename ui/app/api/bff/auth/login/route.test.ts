import { describe, expect, it, vi, beforeEach } from "vitest";

import { POST } from "./route";

function tokenFor(payload: Record<string, unknown>) {
  const encodedPayload = Buffer.from(JSON.stringify(payload), "utf8").toString("base64url");
  return `header.${encodedPayload}.signature`;
}

describe("BFF login route", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("stores Go tokens in HttpOnly cookies and returns a client session", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            code: 200,
            message: "success",
            data: {
              token: {
                access_token: tokenFor({ uid: 1, role: "admin", exp: 4_102_444_800 }),
                token_type: "Bearer",
                expires_in: 900,
                refresh_token: tokenFor({ uid: 1, role: "admin", exp: 4_102_444_800, use: "refresh" })
              }
            }
          }),
          { status: 200, headers: { "content-type": "application/json" } }
        )
      )
    );

    const response = await POST(
      new Request("http://localhost/api/bff/auth/login", {
        method: "POST",
        body: JSON.stringify({ grant_type: "local", account: "admin", password: "password123" })
      })
    );
    const body = await response.json();

    expect(response.status).toBe(200);
    expect(body.data).toMatchObject({ isAuthenticated: true, isAdmin: true, role: "admin" });
    expect(response.headers.getSetCookie().join("\n")).toContain("HttpOnly");
  });
});
