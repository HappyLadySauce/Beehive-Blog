import { describe, expect, it } from "vitest";

import { decodeJwtClaims, isAdminClaims, isExpiredClaims } from "./session";

function tokenFor(payload: Record<string, unknown>) {
  const encodedPayload = Buffer.from(JSON.stringify(payload), "utf8").toString("base64url");
  return `header.${encodedPayload}.signature`;
}

describe("session JWT claims", () => {
  it("decodes role and uid from an access token payload", () => {
    const claims = decodeJwtClaims(tokenFor({ uid: 42, role: "admin", exp: 4_102_444_800, sid: 8, use: "access" }));

    expect(claims).toMatchObject({
      uid: 42,
      role: "admin",
      exp: 4_102_444_800,
      sid: 8,
      use: "access"
    });
    expect(isAdminClaims(claims)).toBe(true);
  });

  it("treats member, malformed, and expired claims as non-admin", () => {
    expect(isAdminClaims(decodeJwtClaims(tokenFor({ uid: 7, role: "member", exp: 4_102_444_800 })))).toBe(false);
    expect(isAdminClaims(decodeJwtClaims("not-a-jwt"))).toBe(false);
    expect(isAdminClaims(decodeJwtClaims(tokenFor({ uid: 9, role: "admin", exp: 1 })))).toBe(false);
  });

  it("detects expired tokens", () => {
    expect(isExpiredClaims({ exp: 1 })).toBe(true);
    expect(isExpiredClaims({ exp: 4_102_444_800 })).toBe(false);
    expect(isExpiredClaims(null)).toBe(true);
  });
});
