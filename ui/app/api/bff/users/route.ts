import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

import type { AuthPayload, CreateUserRequest, CreateUserResponse, ListUsersResponse } from "@/lib/api/types";
import { BffAuthError, forwardGoApi, jsonError, refreshAuthSession, setAuthCookies } from "@/lib/auth/bff";
import { accessCookieName, refreshCookieName } from "@/lib/auth/cookies";

const usersPath = "/users";

type ForwardResult<T> = {
  data: T;
  refreshedAuth?: AuthPayload;
};

export async function GET(request: NextRequest) {
  try {
    const search = request.nextUrl.search;
    const result = await forwardUsersRequest<ListUsersResponse>({ method: "GET" }, search);
    return usersResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function POST(request: Request) {
  try {
    const body = (await request.json()) as CreateUserRequest;
    const result = await forwardUsersRequest<CreateUserResponse>({
      method: "POST",
      body: JSON.stringify(body)
    });
    return usersResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

async function forwardUsersRequest<T>(init: RequestInit, search = ""): Promise<ForwardResult<T>> {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get(accessCookieName)?.value;
  const refreshToken = cookieStore.get(refreshCookieName)?.value;

  if (accessToken) {
    try {
      const data = await forwardGoApi<T>(`${usersPath}${search}`, withBearer(init, accessToken));
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
  const data = await forwardGoApi<T>(`${usersPath}${search}`, withBearer(init, refreshedAuth.token.access_token));
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
