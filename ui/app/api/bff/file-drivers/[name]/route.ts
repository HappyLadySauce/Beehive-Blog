import type { DriverResponse } from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoRequest } from "@/lib/auth/bff-proxy";

export async function GET(_request: Request, { params }: { params: Promise<{ name: string }> }) {
  const { name } = await params;
  try {
    const result = await forwardAuthedGoRequest<DriverResponse>(`/file-drivers/${encodeURIComponent(name)}`, {
      method: "GET"
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
