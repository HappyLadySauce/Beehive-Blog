import { planStatusTransition } from "../content-status";
import { apiFetch } from "./client";
import type {
  ContentDetailResponse,
  ContentStatus,
  CreateContentRequest,
  CreateContentResponse,
  DeleteContentResponse,
  ListContentsRequest,
  ListContentsResponse,
  SetContentCategoriesRequest,
  SetContentTagsRequest,
  TransitionContentStatusRequest,
  UpdateContentRequest
} from "./types";

export function listContents(params: ListContentsRequest = {}) {
  const search = new URLSearchParams();
  if (params.page) search.set("page", String(params.page));
  if (params.page_size) search.set("page_size", String(params.page_size));
  if (params.type) search.set("type", params.type);
  if (params.status) search.set("status", params.status);
  if (params.visibility) search.set("visibility", params.visibility);
  if (params.tag_id) search.set("tag_id", String(params.tag_id));
  if (params.category_id) search.set("category_id", String(params.category_id));
  if (params.search) search.set("search", params.search);
  if (params.slug) search.set("slug", params.slug);
  const query = search.toString();
  return apiFetch<ListContentsResponse>(`/bff/contents${query ? `?${query}` : ""}`, { method: "GET" });
}

export function getContent(id: number) {
  return apiFetch<ContentDetailResponse>(`/bff/contents/${id}`, { method: "GET" });
}

export function createContent(payload: CreateContentRequest) {
  return apiFetch<CreateContentResponse>("/bff/contents", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function updateContent(id: number, payload: UpdateContentRequest) {
  return apiFetch<ContentDetailResponse>(`/bff/contents/${id}`, {
    method: "PATCH",
    body: JSON.stringify(payload)
  });
}

export function deleteContent(id: number) {
  return apiFetch<DeleteContentResponse>(`/bff/contents/${id}`, { method: "DELETE" });
}

export function transitionContentStatus(id: number, payload: TransitionContentStatusRequest) {
  return apiFetch<ContentDetailResponse>(`/bff/contents/${id}/status`, {
    method: "PATCH",
    body: JSON.stringify(payload)
  });
}

/**
 * Applies a multi-step status transition when the target is not directly reachable.
 * transitionContentStatusTo 在目标状态不可直达时按最短路径串行调用状态流转 API。
 */
export async function transitionContentStatusTo(
  id: number,
  from: string,
  to: ContentStatus
): Promise<ContentDetailResponse | null> {
  const steps = planStatusTransition(from, to);
  if (steps.length === 0) {
    if (from === to) {
      return null;
    }
    throw new Error(`invalid status transition from "${from}" to "${to}"`);
  }

  let last: ContentDetailResponse | null = null;
  for (const status of steps) {
    last = await transitionContentStatus(id, { status });
  }
  return last;
}

export function setContentTags(id: number, payload: SetContentTagsRequest) {
  return apiFetch<Record<string, never>>(`/bff/contents/${id}/tags`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export function setContentCategories(id: number, payload: SetContentCategoriesRequest) {
  return apiFetch<Record<string, never>>(`/bff/contents/${id}/categories`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}
