import type { AttachmentPatchRequest, AttachmentResponse, DeleteAttachmentResponse } from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoRequest } from "@/lib/auth/bff-proxy";

type RouteContext = {
  params: Promise<{ id: string }>;
};

export async function GET(_request: Request, context: RouteContext) {
  try {
    const { id } = await context.params;
    const result = await forwardAuthedGoRequest<AttachmentResponse>(`/attachments/${id}`, { method: "GET" });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function PATCH(request: Request, context: RouteContext) {
  try {
    const { id } = await context.params;
    const body = (await request.json()) as AttachmentPatchRequest;
    const result = await forwardAuthedGoRequest<AttachmentResponse>(`/attachments/${id}`, {
      method: "PATCH",
      body: JSON.stringify(body)
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function DELETE(request: Request, context: RouteContext) {
  try {
    const { id } = await context.params;
    const url = new URL(request.url);
    const force = url.searchParams.get("force") === "true";
    const result = await forwardAuthedGoRequest<DeleteAttachmentResponse>(`/attachments/${id}${force ? "?force=true" : ""}`, {
      method: "DELETE"
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
