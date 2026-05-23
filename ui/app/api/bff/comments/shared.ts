import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import type { AuthPayload } from "@/lib/api/types";
import { BffAuthError, forwardGoApi, refreshAuthSession, setAuthCookies } from "@/lib/auth/bff";
import { accessCookieName, refreshCookieName } from "@/lib/auth/cookies";

export type ForwardResult<T> = { data: T; refreshedAuth?: AuthPayload };

export async function forwardAuthed<T>(path: string, init: RequestInit): Promise<ForwardResult<T>> {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get(accessCookieName)?.value;
  const refreshToken = cookieStore.get(refreshCookieName)?.value;
  if (accessToken) {
    try {
      return { data: await forwardGoApi<T>(path, withBearer(init, accessToken)) };
    } catch (error) {
      if (!(error instanceof BffAuthError) || error.status !== 401 || !refreshToken) throw error;
    }
  }
  if (!refreshToken) throw new BffAuthError("Missing authenticated session", 401);
  const refreshedAuth = await refreshAuthSession(refreshToken);
  const data = await forwardGoApi<T>(path, withBearer(init, refreshedAuth.token.access_token));
  return { data, refreshedAuth };
}

export function response<T>(result: ForwardResult<T>) {
  const res = NextResponse.json({ code: 200, message: "success", data: result.data });
  if (result.refreshedAuth) setAuthCookies(res, result.refreshedAuth);
  return res;
}

function withBearer(init: RequestInit, accessToken: string): RequestInit {
  return { ...init, headers: { ...init.headers, authorization: `Bearer ${accessToken}` } };
}
