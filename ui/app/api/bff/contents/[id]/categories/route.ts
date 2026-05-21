import type { CategoryItem, SetContentCategoriesRequest } from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoRequest } from "@/lib/auth/bff-proxy";

type RouteContext = {
  params: Promise<{ id: string }>;
};

export async function GET(_request: Request, context: RouteContext) {
  try {
    const { id } = await context.params;
    const result = await forwardAuthedGoRequest<CategoryItem[]>(`/contents/${id}/categories`, { method: "GET" });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function PUT(request: Request, context: RouteContext) {
  try {
    const { id } = await context.params;
    const body = (await request.json()) as SetContentCategoriesRequest;
    const result = await forwardAuthedGoRequest<Record<string, never>>(`/contents/${id}/categories`, {
      method: "PUT",
      body: JSON.stringify(body)
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
