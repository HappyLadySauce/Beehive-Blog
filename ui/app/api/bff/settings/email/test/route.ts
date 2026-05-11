import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import type { AuthPayload, SettingsEmailTestResponse } from "@/lib/api/types";
import { BffAuthError, forwardGoApi, jsonError, refreshAuthSession, setAuthCookies } from "@/lib/auth/bff";
import { accessCookieName, refreshCookieName } from "@/lib/auth/cookies";

type EmailTestForwardResult = {
  data: SettingsEmailTestResponse;
  refreshedAuth?: AuthPayload;
};

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const result = await forwardEmailTestRequest<SettingsEmailTestResponse>({
      method: "POST",
      body: JSON.stringify(body)
    });
    return emailTestResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

async function forwardEmailTestRequest<T extends SettingsEmailTestResponse>(init: RequestInit): Promise<EmailTestForwardResult> {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get(accessCookieName)?.value;
  const refreshToken = cookieStore.get(refreshCookieName)?.value;

  if (accessToken) {
    try {
      const data = await forwardGoApi<T>("/settings/email/test", withBearer(init, accessToken));
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
  const data = await forwardGoApi<T>("/settings/email/test", withBearer(init, refreshedAuth.token.access_token));
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

function emailTestResponse(result: EmailTestForwardResult) {
  const response = NextResponse.json({ code: 200, message: "success", data: result.data });
  if (result.refreshedAuth) {
    setAuthCookies(response, result.refreshedAuth);
  }
  return response;
}
