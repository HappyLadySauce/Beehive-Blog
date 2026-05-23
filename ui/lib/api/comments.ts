import { apiFetch } from "./client";
import type { ListAdminCommentsRequest, ListAdminCommentsResponse, AdminCommentItem, UpdateCommentStatusRequest } from "./types";

export function listAdminComments(params: ListAdminCommentsRequest = {}) {
  const search = new URLSearchParams();
  if (params.page) search.set("page", String(params.page));
  if (params.page_size) search.set("page_size", String(params.page_size));
  if (params.status) search.set("status", params.status);
  if (params.content_id) search.set("content_id", String(params.content_id));
  if (params.search) search.set("search", params.search);
  const query = search.toString();
  return apiFetch<ListAdminCommentsResponse>(`/bff/comments${query ? `?${query}` : ""}`, { method: "GET" });
}

export function updateAdminCommentStatus(id: number, payload: UpdateCommentStatusRequest) {
  return apiFetch<AdminCommentItem>(`/bff/comments/${id}/status`, {
    method: "PATCH",
    body: JSON.stringify(payload)
  });
}

export function deleteAdminComment(id: number) {
  return apiFetch<Record<string, never>>(`/bff/comments/${id}`, { method: "DELETE" });
}
