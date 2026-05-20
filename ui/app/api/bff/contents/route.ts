import type { CreateContentRequest, CreateContentResponse, ListContentsResponse } from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoRequest } from "@/lib/auth/bff-proxy";

export async function GET(request: Request) {
  try {
    const url = new URL(request.url);
    const result = await forwardAuthedGoRequest<ListContentsResponse>(`/contents${url.search}`, { method: "GET" });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function POST(request: Request) {
  try {
    const body = (await request.json()) as CreateContentRequest;
    const result = await forwardAuthedGoRequest<CreateContentResponse>("/contents", {
      method: "POST",
      body: JSON.stringify(body)
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
