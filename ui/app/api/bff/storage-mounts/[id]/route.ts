import type {
  DeleteStorageMountResponse,
  StorageMountPatchRequest,
  StorageMountResponse
} from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoRequest } from "@/lib/auth/bff-proxy";

export async function GET(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  try {
    const result = await forwardAuthedGoRequest<StorageMountResponse>(`/storage-mounts/${id}`, { method: "GET" });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function PATCH(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  try {
    const body = (await request.json()) as StorageMountPatchRequest;
    const result = await forwardAuthedGoRequest<StorageMountResponse>(`/storage-mounts/${id}`, {
      method: "PATCH",
      body: JSON.stringify(body)
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function DELETE(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  try {
    const result = await forwardAuthedGoRequest<DeleteStorageMountResponse>(`/storage-mounts/${id}`, {
      method: "DELETE"
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
