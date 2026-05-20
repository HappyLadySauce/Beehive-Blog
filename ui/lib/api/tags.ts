import { apiFetch } from "./client";
import type {
  CreateTagRequest,
  CreateTagResponse,
  DeleteTagResponse,
  ListTagsRequest,
  ListTagsResponse,
  TagDetailResponse,
  UpdateTagRequest
} from "./types";

export function listTags(params: ListTagsRequest = {}) {
  const search = new URLSearchParams();
  if (params.page) search.set("page", String(params.page));
  if (params.page_size) search.set("page_size", String(params.page_size));
  if (params.status) search.set("status", params.status);
  if (params.search) search.set("search", params.search);
  const query = search.toString();
  return apiFetch<ListTagsResponse>(`/bff/tags${query ? `?${query}` : ""}`, { method: "GET" });
}

export function getTag(id: number) {
  return apiFetch<TagDetailResponse>(`/bff/tags/${id}`, { method: "GET" });
}

export function createTag(payload: CreateTagRequest) {
  return apiFetch<CreateTagResponse>("/bff/tags", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function updateTag(id: number, payload: UpdateTagRequest) {
  return apiFetch<TagDetailResponse>(`/bff/tags/${id}`, {
    method: "PATCH",
    body: JSON.stringify(payload)
  });
}

export function deleteTag(id: number) {
  return apiFetch<DeleteTagResponse>(`/bff/tags/${id}`, { method: "DELETE" });
}
