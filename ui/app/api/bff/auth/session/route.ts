import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { clearAuthCookies, jsonError, refreshAuthSession, sessionFromAuthPayload, setAuthCookies } from "@/lib/auth/bff";
import { accessCookieName, refreshCookieName } from "@/lib/auth/cookies";
import { decodeJwtClaims, isExpiredClaims, sessionFromClaims } from "@/lib/auth/session";

export async function GET() {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get(accessCookieName)?.value;
  const refreshToken = cookieStore.get(refreshCookieName)?.value;
  const accessClaims = accessToken ? decodeJwtClaims(accessToken) : null;

  if (accessClaims && !isExpiredClaims(accessClaims)) {
    return NextResponse.json({ code: 200, message: "success", data: sessionFromClaims(accessClaims) });
  }

  if (!refreshToken) {
    const response = NextResponse.json({ code: 200, message: "success", data: sessionFromClaims(null) });
    clearAuthCookies(response);
    return response;
  }

  try {
    const authPayload = await refreshAuthSession(refreshToken);
    const session = sessionFromAuthPayload(authPayload);
    const response = NextResponse.json({ code: 200, message: "success", data: session });
    setAuthCookies(response, authPayload);
    return response;
  } catch (error) {
    const response = jsonError(error);
    clearAuthCookies(response);
    return response;
  }
}
