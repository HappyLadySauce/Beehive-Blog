import type { AttachmentBatchUploadResponse } from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoMultipartRequest } from "@/lib/auth/bff-proxy";

export async function POST(request: Request) {
  try {
    const formData = await request.formData();
    const result = await forwardAuthedGoMultipartRequest<AttachmentBatchUploadResponse>("/attachments/batch", {
      method: "POST",
      body: formData
    });
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
