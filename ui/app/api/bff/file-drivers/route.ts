import type { DriverListResponse } from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoRequest } from "@/lib/auth/bff-proxy";

export async function GET() {
  try {
    const result = await forwardAuthedGoRequest<DriverListResponse>("/file-drivers", { method: "GET" });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
