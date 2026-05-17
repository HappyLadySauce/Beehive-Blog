import { jsonError } from "@/lib/auth/bff";
import { bffRawResponse, forwardAuthedGoRawResponse } from "@/lib/auth/bff-proxy";

type RouteContext = {
  params: Promise<{ id: string }>;
};

export async function GET(_request: Request, context: RouteContext) {
  try {
    const { id } = await context.params;
    const result = await forwardAuthedGoRawResponse(`/attachments/${id}/content`, { method: "GET" });
    return bffRawResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
