import type { AdminCommentItem } from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { forwardAuthed, response } from "../../shared";

type RouteContext = {
  params: Promise<{ id: string }>;
};

export async function PATCH(request: Request, context: RouteContext) {
  try {
    const { id } = await context.params;
    const body = await request.json();
    const result = await forwardAuthed<AdminCommentItem>(`/comments/${id}/status`, {
      method: "PATCH",
      body: JSON.stringify(body)
    });
    return response(result);
  } catch (error) {
    return jsonError(error);
  }
}
