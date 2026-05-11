import { NextResponse } from "next/server";

import type { AuthPayload, LoginRequest } from "@/lib/api/types";
import { forwardGoApi, jsonError, sessionFromAuthPayload, setAuthCookies } from "@/lib/auth/bff";

export async function POST(request: Request) {
  try {
    const payload = (await request.json()) as LoginRequest;
    const authPayload = await forwardGoApi<AuthPayload>("/auth/login", {
      method: "POST",
      body: JSON.stringify(payload)
    });
    const session = sessionFromAuthPayload(authPayload);
    const response = NextResponse.json({ code: 200, message: "success", data: session });
    setAuthCookies(response, authPayload);
    return response;
  } catch (error) {
    return jsonError(error);
  }
}
