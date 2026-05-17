import type {
  AttachmentCategoryCreateRequest,
  AttachmentCategoryListResponse,
  AttachmentCategoryResponse
} from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoRequest } from "@/lib/auth/bff-proxy";

export async function GET() {
  try {
    const result = await forwardAuthedGoRequest<AttachmentCategoryListResponse>("/attachment/categories", { method: "GET" });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function POST(request: Request) {
  try {
    const body = (await request.json()) as AttachmentCategoryCreateRequest;
    const result = await forwardAuthedGoRequest<AttachmentCategoryResponse>("/attachment/categories", {
      method: "POST",
      body: JSON.stringify(body)
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
