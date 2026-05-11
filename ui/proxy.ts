import { NextRequest, NextResponse } from "next/server";

import type { AuthPayload, BaseResponse } from "@/lib/api/types";
import type { VerifiedAccessSession } from "@/lib/auth/bff";
import { accessCookieName, refreshCookieName, secureCookieEnabled } from "@/lib/auth/cookies";
import { decodeJwtClaims, isAdminClaims, isExpiredClaims } from "@/lib/auth/session";

const goApiBaseUrl = process.env.BEEHIVE_API_BASE_URL ?? "http://localhost:8080";
const fallbackRefreshMaxAge = 60 * 60 * 24 * 30;

export async function proxy(request: NextRequest) {
  const accessToken = request.cookies.get(accessCookieName)?.value;

  if (accessToken) {
    const verifiedSession = await verifyAccessFromGo(accessToken).catch((error) => {
      console.error("Studio proxy access verification failed", error);
      return null;
    });
    if (isAdminClaims(verifiedSession)) {
      return NextResponse.next();
    }
    if (verifiedSession?.role && !isExpiredClaims(verifiedSession)) {
      return NextResponse.redirect(new URL("/", request.url));
    }
  }

  const refreshToken = request.cookies.get(refreshCookieName)?.value;
  if (refreshToken) {
    const refreshed = await refreshFromGo(refreshToken).catch((error) => {
      console.error("Studio proxy refresh failed", error);
      return null;
    });

    if (refreshed) {
      const refreshedClaims = decodeJwtClaims(refreshed.token.access_token);
      if (isAdminClaims(refreshedClaims)) {
        const response = NextResponse.next();
        setProxyAuthCookies(response, refreshed);
        return response;
      }
      if (refreshedClaims?.role && !isExpiredClaims(refreshedClaims)) {
        const response = NextResponse.redirect(new URL("/", request.url));
        setProxyAuthCookies(response, refreshed);
        return response;
      }
    }
  }

  const target = new URL("/login", request.url);
  target.searchParams.set("next", request.nextUrl.pathname);
  const response = NextResponse.redirect(target);
  clearProxyAuthCookies(response);
  return response;
}

async function verifyAccessFromGo(accessToken: string) {
  const response = await fetch(`${goApiBaseUrl}/api/v1/auth/session`, {
    method: "GET",
    headers: {
      authorization: `Bearer ${accessToken}`
    },
    cache: "no-store"
  });

  const envelope = (await response.json()) as BaseResponse<VerifiedAccessSession>;
  if (!response.ok || envelope.code < 200 || envelope.code >= 300) {
    return null;
  }
  return envelope.data;
}

async function refreshFromGo(refreshToken: string) {
  const response = await fetch(`${goApiBaseUrl}/api/v1/auth/refresh`, {
    method: "POST",
    headers: {
      "content-type": "application/json"
    },
    body: JSON.stringify({ refresh_token: refreshToken }),
    cache: "no-store"
  });

  const envelope = (await response.json()) as BaseResponse<AuthPayload>;
  if (!response.ok || envelope.code < 200 || envelope.code >= 300) {
    return null;
  }
  return envelope.data;
}

function setProxyAuthCookies(response: NextResponse, payload: AuthPayload) {
  const refreshClaims = payload.token.refresh_token ? decodeJwtClaims(payload.token.refresh_token) : null;
  response.cookies.set(accessCookieName, payload.token.access_token, {
    httpOnly: true,
    maxAge: Math.max(1, payload.token.expires_in),
    path: "/",
    sameSite: "lax",
    secure: secureCookieEnabled()
  });
  if (payload.token.refresh_token) {
    response.cookies.set(refreshCookieName, payload.token.refresh_token, {
      httpOnly: true,
      maxAge: refreshClaims?.exp ? Math.max(1, Math.floor(refreshClaims.exp - Date.now() / 1000)) : fallbackRefreshMaxAge,
      path: "/",
      sameSite: "lax",
      secure: secureCookieEnabled()
    });
  }
}

function clearProxyAuthCookies(response: NextResponse) {
  response.cookies.set(accessCookieName, "", { maxAge: 0, path: "/" });
  response.cookies.set(refreshCookieName, "", { maxAge: 0, path: "/" });
}

export const config = {
  matcher: ["/studio/:path*"]
};
