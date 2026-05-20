import type { DeleteTagResponse, TagDetailResponse, UpdateTagRequest } from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoRequest } from "@/lib/auth/bff-proxy";

type RouteContext = {
  params: Promise<{ id: string }>;
};

export async function GET(_request: Request, context: RouteContext) {
  try {
    const { id } = await context.params;
    const result = await forwardAuthedGoRequest<TagDetailResponse>(`/tags/${id}`, { method: "GET" });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function PATCH(request: Request, context: RouteContext) {
  try {
    const { id } = await context.params;
    const body = (await request.json()) as UpdateTagRequest;
    const result = await forwardAuthedGoRequest<TagDetailResponse>(`/tags/${id}`, {
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
    const result = await forwardAuthedGoRequest<DeleteTagResponse>(`/tags/${id}`, { method: "DELETE" });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
