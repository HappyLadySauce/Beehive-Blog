import type { AttachmentPresignRequest, AttachmentPresignResponse } from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoRequest } from "@/lib/auth/bff-proxy";

export async function POST(request: Request) {
  try {
    const body = (await request.json()) as AttachmentPresignRequest;
    const result = await forwardAuthedGoRequest<AttachmentPresignResponse>("/attachments/upload-url", {
      method: "POST",
      body: JSON.stringify(body)
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
