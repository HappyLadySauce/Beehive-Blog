import type { CreateTagRequest, CreateTagResponse, ListTagsResponse } from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoRequest } from "@/lib/auth/bff-proxy";

export async function GET(request: Request) {
  try {
    const url = new URL(request.url);
    const result = await forwardAuthedGoRequest<ListTagsResponse>(`/tags${url.search}`, { method: "GET" });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function POST(request: Request) {
  try {
    const body = (await request.json()) as CreateTagRequest;
    const result = await forwardAuthedGoRequest<CreateTagResponse>("/tags", {
      method: "POST",
      body: JSON.stringify(body)
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
