import type { ListAdminCommentsResponse } from "@/lib/api/types";
import { jsonError } from "@/lib/auth/bff";
import { forwardAuthed, response } from "./shared";

export async function GET(request: Request) {
  try {
    const url = new URL(request.url);
    const result = await forwardAuthed<ListAdminCommentsResponse>(`/comments${url.search}`, { method: "GET" });
    return response(result);
  } catch (error) {
    return jsonError(error);
  }
}
