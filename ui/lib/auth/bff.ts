import { NextResponse } from "next/server";

import type { AuthPayload, BaseResponse } from "@/lib/api/types";
import { decodeJwtClaims, sessionFromClaims } from "@/lib/auth/session";
import { accessCookieName, refreshCookieName, secureCookieEnabled } from "@/lib/auth/cookies";

const goApiBaseUrl = process.env.BEEHIVE_API_BASE_URL ?? "http://localhost:8080";
const fallbackRefreshMaxAge = 60 * 60 * 24 * 30;

export class BffAuthError extends Error {
  readonly status: number;
  readonly code: number;

  constructor(message: string, status: number, code = status) {
    super(message);
    this.name = "BffAuthError";
    this.status = status;
    this.code = code;
  }
}

export async function forwardGoApi<T>(path: string, init: RequestInit): Promise<T> {
  const response = await fetch(`${goApiBaseUrl}/api/v1${path}`, {
    ...init,
    headers: {
      "content-type": "application/json",
      ...init.headers
    },
    cache: "no-store"
  });

  const contentType = response.headers.get("content-type") ?? "";
  if (!contentType.includes("application/json")) {
    throw new BffAuthError("API returned an invalid response", response.status, response.status);
  }

  let envelope: BaseResponse<T>;
  try {
    envelope = (await response.json()) as BaseResponse<T>;
  } catch {
    throw new BffAuthError("API returned an invalid response", response.status, response.status);
  }

  if (!response.ok || envelope.code < 200 || envelope.code >= 300) {
    throw new BffAuthError(envelope.message || "Request failed", response.status, envelope.code);
  }

  return envelope.data;
}

export function jsonError(error: unknown) {
  if (error instanceof BffAuthError) {
    return NextResponse.json({ code: error.code, message: error.message, data: null }, { status: error.status });
  }

  console.error("BFF auth request failed", error);
  return NextResponse.json({ code: 500, message: "Authentication service is unavailable", data: null }, { status: 500 });
}

export function setAuthCookies(response: NextResponse, payload: AuthPayload) {
  const accessClaims = decodeJwtClaims(payload.token.access_token);
  const refreshClaims = payload.token.refresh_token ? decodeJwtClaims(payload.token.refresh_token) : null;
  const accessMaxAge = Math.max(1, payload.token.expires_in);
  const refreshMaxAge = refreshClaims?.exp
    ? Math.max(1, Math.floor(refreshClaims.exp - Date.now() / 1000))
    : fallbackRefreshMaxAge;

  response.cookies.set(accessCookieName, payload.token.access_token, {
    httpOnly: true,
    maxAge: accessMaxAge,
    path: "/",
    sameSite: "lax",
    secure: secureCookieEnabled()
  });

  if (payload.token.refresh_token) {
    response.cookies.set(refreshCookieName, payload.token.refresh_token, {
      httpOnly: true,
      maxAge: refreshMaxAge,
      path: "/",
      sameSite: "lax",
      secure: secureCookieEnabled()
    });
  }

  return sessionFromClaims(accessClaims);
}

export function sessionFromAuthPayload(payload: AuthPayload) {
  return sessionFromClaims(decodeJwtClaims(payload.token.access_token));
}

export function clearAuthCookies(response: NextResponse) {
  response.cookies.set(accessCookieName, "", {
    httpOnly: true,
    maxAge: 0,
    path: "/",
    sameSite: "lax",
    secure: secureCookieEnabled()
  });
  response.cookies.set(refreshCookieName, "", {
    httpOnly: true,
    maxAge: 0,
    path: "/",
    sameSite: "lax",
    secure: secureCookieEnabled()
  });
}

export async function refreshAuthSession(refreshToken: string) {
  return forwardGoApi<AuthPayload>("/auth/refresh", {
    method: "POST",
    body: JSON.stringify({ refresh_token: refreshToken })
  });
}
