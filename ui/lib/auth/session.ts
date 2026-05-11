export type JwtClaims = {
  uid?: number;
  role?: string;
  exp?: number;
  sid?: number;
  use?: string;
};

export type ClientSession = {
  isAuthenticated: boolean;
  claims: JwtClaims | null;
  role?: string;
  isAdmin: boolean;
  isTokenExpired: boolean;
};

export function decodeJwtClaims(token: string): JwtClaims | null {
  const payload = token.split(".")[1];
  if (!payload) return null;

  try {
    const normalized = payload.replace(/-/g, "+").replace(/_/g, "/");
    const padded = normalized.padEnd(normalized.length + ((4 - (normalized.length % 4)) % 4), "=");
    const json =
      typeof atob === "function"
        ? decodeURIComponent(
            Array.from(atob(padded), (char) => `%${char.charCodeAt(0).toString(16).padStart(2, "0")}`).join("")
          )
        : Buffer.from(padded, "base64").toString("utf8");
    const claims = JSON.parse(json) as JwtClaims;
    return typeof claims === "object" && claims !== null ? claims : null;
  } catch {
    return null;
  }
}

export function isExpiredClaims(claims: JwtClaims | null | undefined) {
  if (!claims?.exp) return true;
  return claims.exp * 1000 <= Date.now();
}

export function isAdminClaims(claims: JwtClaims | null | undefined) {
  return claims?.role === "admin" && !isExpiredClaims(claims);
}

export function sessionFromClaims(claims: JwtClaims | null): ClientSession {
  const isTokenExpired = isExpiredClaims(claims);
  const isAuthenticated = Boolean(claims?.role && !isTokenExpired);

  return {
    isAuthenticated,
    claims: isAuthenticated ? claims : null,
    role: isAuthenticated ? claims?.role : undefined,
    isAdmin: isAdminClaims(claims),
    isTokenExpired
  };
}
