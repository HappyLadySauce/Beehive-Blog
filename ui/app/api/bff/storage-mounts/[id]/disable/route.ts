import type { StorageMountResponse } from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoRequest } from "@/lib/auth/bff-proxy";

export async function POST(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  try {
    const result = await forwardAuthedGoRequest<StorageMountResponse>(`/storage-mounts/${id}/disable`, {
      method: "POST"
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
