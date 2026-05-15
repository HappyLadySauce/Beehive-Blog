import type { StorageMountCreateRequest, StorageMountListResponse, StorageMountResponse } from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoRequest } from "@/lib/auth/bff-proxy";

export async function GET() {
  try {
    const result = await forwardAuthedGoRequest<StorageMountListResponse>("/storage-mounts", { method: "GET" });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function POST(request: Request) {
  try {
    const body = (await request.json()) as StorageMountCreateRequest;
    const result = await forwardAuthedGoRequest<StorageMountResponse>("/storage-mounts", {
      method: "POST",
      body: JSON.stringify(body)
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
