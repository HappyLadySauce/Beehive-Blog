import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { forwardGoApi, jsonError, clearAuthCookies } from "@/lib/auth/bff";
import { accessCookieName } from "@/lib/auth/cookies";

export async function POST() {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get(accessCookieName)?.value;

  try {
    if (accessToken) {
      await forwardGoApi<unknown>("/auth/logout", {
        method: "POST",
        headers: {
          authorization: `Bearer ${accessToken}`
        }
      });
    }
    const response = NextResponse.json({ code: 200, message: "success", data: null });
    clearAuthCookies(response);
    return response;
  } catch (error) {
    const response = jsonError(error);
    clearAuthCookies(response);
    return response;
  }
}
