import type { AttachmentReferenceListResponse } from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoRequest } from "@/lib/auth/bff-proxy";

export async function GET(request: Request) {
  try {
    const url = new URL(request.url);
    const result = await forwardAuthedGoRequest<AttachmentReferenceListResponse>(`/attachments/references${url.search}`, { method: "GET" });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
