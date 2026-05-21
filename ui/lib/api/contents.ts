import { apiFetch } from "./client";
import type {
  ContentDetailResponse,
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
