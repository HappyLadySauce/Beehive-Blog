import type { CreateVersionRequest, CreateVersionResponse, ListVersionsResponse } from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoRequest } from "@/lib/auth/bff-proxy";

type RouteContext = {
  params: Promise<{ id: string }>;
};

export async function GET(request: Request, context: RouteContext) {
  try {
    const { id } = await context.params;
    const result = await forwardAuthedGoRequest<ListVersionsResponse>(`/contents/${id}/versions`, {
      method: "GET"
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function POST(request: Request, context: RouteContext) {
  try {
    const { id } = await context.params;
    const body = (await request.json()) as CreateVersionRequest;
    const result = await forwardAuthedGoRequest<CreateVersionResponse>(`/contents/${id}/versions`, {
      method: "POST",
      body: JSON.stringify(body)
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
