import type { AttachmentListResponse, AttachmentResponse } from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoMultipartRequest, forwardAuthedGoRequest } from "@/lib/auth/bff-proxy";

export async function GET(request: Request) {
  try {
    const url = new URL(request.url);
    const result = await forwardAuthedGoRequest<AttachmentListResponse>(`/attachments${url.search}`, { method: "GET" });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}

export async function POST(request: Request) {
  try {
    const formData = await request.formData();
    const result = await forwardAuthedGoMultipartRequest<AttachmentResponse>("/attachments", {
      method: "POST",
      body: formData
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
