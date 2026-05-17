import type {
  AttachmentCategoryPatchRequest,
  AttachmentCategoryResponse,
  DeleteAttachmentCategoryResponse
} from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoRequest } from "@/lib/auth/bff-proxy";

type RouteContext = {
  params: Promise<{ id: string }>;
};

export async function GET(_request: Request, context: RouteContext) {
  try {
    const { id } = await context.params;
    const result = await forwardAuthedGoRequest<AttachmentCategoryResponse>(`/attachment/categories/${id}`, { method: "GET" });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function PATCH(request: Request, context: RouteContext) {
  try {
    const { id } = await context.params;
    const body = (await request.json()) as AttachmentCategoryPatchRequest;
    const result = await forwardAuthedGoRequest<AttachmentCategoryResponse>(`/attachment/categories/${id}`, {
      method: "PATCH",
      body: JSON.stringify(body)
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function DELETE(_request: Request, context: RouteContext) {
  try {
    const { id } = await context.params;
    const result = await forwardAuthedGoRequest<DeleteAttachmentCategoryResponse>(`/attachment/categories/${id}`, {
      method: "DELETE"
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
