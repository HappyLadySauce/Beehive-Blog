import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import type { AuthPayload } from "@/lib/api/types";
import { BffAuthError, forwardGoApi, goApiUrl, refreshAuthSession, setAuthCookies } from "@/lib/auth/bff";
import { accessCookieName, refreshCookieName } from "@/lib/auth/cookies";

export type BffForwardResult<T> = {
  data: T;
  refreshedAuth?: AuthPayload;
};

export type BffRawForwardResult = {
  response: Response;
  refreshedAuth?: AuthPayload;
};

export async function forwardAuthedGoRequest<T>(path: string, init: RequestInit): Promise<BffForwardResult<T>> {
  return forwardAuthedGoRequestWithOptions(path, init);
}

export async function forwardAuthedGoMultipartRequest<T>(path: string, init: RequestInit): Promise<BffForwardResult<T>> {
  return forwardAuthedGoRequestWithOptions(path, init, { defaultContentType: false });
}

async function forwardAuthedGoRequestWithOptions<T>(
  path: string,
  init: RequestInit,
  options: { defaultContentType?: boolean } = {}
): Promise<BffForwardResult<T>> {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get(accessCookieName)?.value;
  const refreshToken = cookieStore.get(refreshCookieName)?.value;

  if (accessToken) {
    try {
      const data = await forwardGoApi<T>(path, withBearer(init, accessToken), options);
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
  const data = await forwardGoApi<T>(path, withBearer(init, refreshedAuth.token.access_token), options);
  return { data, refreshedAuth };
}

export function bffJsonResponse<T>(result: BffForwardResult<T>) {
  const response = NextResponse.json({ code: 200, message: "success", data: result.data });
  if (result.refreshedAuth) {
    setAuthCookies(response, result.refreshedAuth);
  }
  return response;
}

export async function forwardAuthedGoRawResponse(path: string, init: RequestInit): Promise<BffRawForwardResult> {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get(accessCookieName)?.value;
  const refreshToken = cookieStore.get(refreshCookieName)?.value;

  if (accessToken) {
    const response = await fetch(goApiUrl(path), {
      ...withBearer(init, accessToken),
      cache: "no-store",
      redirect: "manual"
    });
    if (response.status !== 401 || !refreshToken) {
      await assertRawForwardSuccess(response);
      return { response };
    }
  }

  if (!refreshToken) {
    throw new BffAuthError("Missing authenticated session", 401);
  }

  const refreshedAuth = await refreshAuthSession(refreshToken);
  const response = await fetch(goApiUrl(path), {
    ...withBearer(init, refreshedAuth.token.access_token),
    cache: "no-store",
    redirect: "manual"
  });
  await assertRawForwardSuccess(response);
  return { response, refreshedAuth };
}

export function bffRawResponse(result: BffRawForwardResult) {
  const headers = new Headers();
  const contentType = result.response.headers.get("content-type");
  const contentDisposition = result.response.headers.get("content-disposition");
  const location = result.response.headers.get("location");

  if (contentType) headers.set("content-type", contentType);
  if (contentDisposition) headers.set("content-disposition", contentDisposition);
  if (location && result.response.status >= 300 && result.response.status < 400) {
    headers.set("location", location);
  }

  const response =
    result.response.status >= 300 && result.response.status < 400
      ? new NextResponse(null, { status: result.response.status, headers })
      : new NextResponse(result.response.body, { status: result.response.status, headers });
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

async function assertRawForwardSuccess(response: Response) {
  if (response.ok || (response.status >= 300 && response.status < 400)) {
    return;
  }
  const contentType = response.headers.get("content-type") ?? "";
  if (contentType.includes("application/json")) {
    try {
      const envelope = (await response.clone().json()) as { code?: number; message?: string };
      throw new BffAuthError(envelope.message || "Request failed", response.status, envelope.code ?? response.status);
    } catch (error) {
      if (error instanceof BffAuthError) throw error;
    }
  }
  throw new BffAuthError("Attachment content is unavailable", response.status, response.status);
}
