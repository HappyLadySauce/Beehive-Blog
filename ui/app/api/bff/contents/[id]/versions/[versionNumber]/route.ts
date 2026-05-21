import { jsonError } from "@/lib/auth/bff";
import { bffJsonResponse, forwardAuthedGoRequest } from "@/lib/auth/bff-proxy";

type RouteContext = {
  params: Promise<{ id: string; versionNumber: string }>;
};

export async function DELETE(request: Request, context: RouteContext) {
  try {
    const { id, versionNumber } = await context.params;
    const result = await forwardAuthedGoRequest<Record<string, never>>(
      `/contents/${id}/versions/${versionNumber}`,
      { method: "DELETE" }
    );
    return bffJsonResponse(result);
  } catch (error) {
    return jsonError(error);
  }
}
