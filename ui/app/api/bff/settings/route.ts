import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import type { AuthPayload, SettingsResponse } from "@/lib/api/types";
import { BffAuthError, forwardGoApi, jsonError, refreshAuthSession, setAuthCookies } from "@/lib/auth/bff";
import { accessCookieName, refreshCookieName } from "@/lib/auth/cookies";

type SettingsForwardResult = {
  data: SettingsResponse;
  refreshedAuth?: AuthPayload;
};

const settingsEmailPath = "/settings/email";

export async function GET() {
  try {
    const result = await forwardSettingsRequest<SettingsResponse>({ method: "GET" });
    return settingsResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function PATCH(request: Request) {
  try {
    const body = await request.json();
    const result = await forwardSettingsRequest<SettingsResponse>({
      method: "PATCH",
      body: JSON.stringify(body)
    });
    return settingsResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

async function forwardSettingsRequest<T extends SettingsResponse>(init: RequestInit): Promise<SettingsForwardResult> {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get(accessCookieName)?.value;
  const refreshToken = cookieStore.get(refreshCookieName)?.value;

  if (accessToken) {
    try {
      const data = await forwardGoApi<T>(settingsEmailPath, withBearer(init, accessToken));
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
  const data = await forwardGoApi<T>(settingsEmailPath, withBearer(init, refreshedAuth.token.access_token));
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

function settingsResponse(result: SettingsForwardResult) {
  const response = NextResponse.json({ code: 200, message: "success", data: result.data });
  if (result.refreshedAuth) {
    setAuthCookies(response, result.refreshedAuth);
  }
  return response;
}
