import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { jsonError, refreshAuthSession, sessionFromAuthPayload, setAuthCookies } from "@/lib/auth/bff";
import { refreshCookieName } from "@/lib/auth/cookies";

export async function POST() {
  const cookieStore = await cookies();
  const refreshToken = cookieStore.get(refreshCookieName)?.value;

  if (!refreshToken) {
    return NextResponse.json({ code: 401, message: "Missing refresh session", data: null }, { status: 401 });
  }

  try {
    const authPayload = await refreshAuthSession(refreshToken);
    const session = sessionFromAuthPayload(authPayload);
    const response = NextResponse.json({ code: 200, message: "success", data: session });
    setAuthCookies(response, authPayload);
    return response;
  } catch (error) {
    return jsonError(error);
  }
}
