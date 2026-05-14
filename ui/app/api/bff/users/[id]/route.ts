import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import type { AuthPayload, UpdateUserRequest, UserDetailResponse } from "@/lib/api/types";
import { BffAuthError, forwardGoApi, jsonError, refreshAuthSession, setAuthCookies } from "@/lib/auth/bff";
import { accessCookieName, refreshCookieName } from "@/lib/auth/cookies";

type ForwardResult<T> = {
  data: T;
  refreshedAuth?: AuthPayload;
};

type DeleteUserResponse = Record<string, never>;

export async function GET(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  try {
    const result = await forwardUsersRequest<UserDetailResponse>({ method: "GET" }, `/${id}`);
    return usersResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function PATCH(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  try {
    const body = (await request.json()) as UpdateUserRequest;
    const result = await forwardUsersRequest<UserDetailResponse>({
      method: "PATCH",
      body: JSON.stringify(body)
    }, `/${id}`);
    return usersResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function DELETE(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  try {
    const result = await forwardUsersRequest<DeleteUserResponse>({ method: "DELETE" }, `/${id}`);
    return usersResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

async function forwardUsersRequest<T>(init: RequestInit, suffix = ""): Promise<ForwardResult<T>> {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get(accessCookieName)?.value;
  const refreshToken = cookieStore.get(refreshCookieName)?.value;

  if (accessToken) {
    try {
      const data = await forwardGoApi<T>(`/users${suffix}`, withBearer(init, accessToken));
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
  const data = await forwardGoApi<T>(`/users${suffix}`, withBearer(init, refreshedAuth.token.access_token));
  return { data, refreshedAuth };
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

function usersResponse<T>(result: ForwardResult<T>) {
  const response = NextResponse.json({ code: 200, message: "success", data: result.data });
  if (result.refreshedAuth) {
    setAuthCookies(response, result.refreshedAuth);
  }
  return response;
}
