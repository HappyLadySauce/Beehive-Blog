import type { DeleteContentResponse } from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { forwardAuthed, response } from "../shared";

type RouteContext = {
  params: Promise<{ id: string }>;
};

export async function DELETE(_request: Request, context: RouteContext) {
  try {
    const { id } = await context.params;
    const result = await forwardAuthed<DeleteContentResponse>(`/comments/${id}`, { method: "DELETE" });
    return response(result);
  } catch (error) {
    return jsonError(error);
  }
}
