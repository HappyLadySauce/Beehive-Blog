import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import type { AuthPayload } from "@/lib/api/types";
import { BffAuthError, forwardGoApi, refreshAuthSession, setAuthCookies } from "@/lib/auth/bff";
import { accessCookieName, refreshCookieName } from "@/lib/auth/cookies";

export type BffForwardResult<T> = {
  data: T;
  refreshedAuth?: AuthPayload;
};

export async function forwardAuthedGoRequest<T>(path: string, init: RequestInit): Promise<BffForwardResult<T>> {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get(accessCookieName)?.value;
  const refreshToken = cookieStore.get(refreshCookieName)?.value;

  if (accessToken) {
    try {
      const data = await forwardGoApi<T>(path, withBearer(init, accessToken));
      return { data };
    } catch (error) {
      if (!(error instanceof BffAuthError) || error.status !== 401 || !refreshToken) {
        throw error;
      }
    }
  }

  if (!refreshToken) {
    throw new BffAuthError("Missing authenticated session", 401);
  }

  const refreshedAuth = await refreshAuthSession(refreshToken);
  const data = await forwardGoApi<T>(path, withBearer(init, refreshedAuth.token.access_token));
  return { data, refreshedAuth };
}

export function bffJsonResponse<T>(result: BffForwardResult<T>) {
  const response = NextResponse.json({ code: 200, message: "success", data: result.data });
  if (result.refreshedAuth) {
    setAuthCookies(response, result.refreshedAuth);
  }
  return response;
}

function withBearer(init: RequestInit, accessToken: string): RequestInit {
  return {
    ...init,
    headers: {
      ...init.headers,
      authorization: `Bearer ${accessToken}`
    }
  };
}
