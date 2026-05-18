import { apiFetch } from "./client";
import type {
  AttachmentCategoryCreateRequest,
  AttachmentCategoryListResponse,
  AttachmentCategoryPatchRequest,
  AttachmentCategoryReplaceRequest,
  AttachmentCategoryResponse,
  AttachmentCompleteRequest,
  AttachmentListRequest,
  AttachmentListResponse,
  AttachmentPatchRequest,
  AttachmentPresignRequest,
  AttachmentPresignResponse,
  AttachmentReferenceListResponse,
  AttachmentResponse,
  DeleteAttachmentCategoryResponse,
  DeleteAttachmentResponse
} from "./types";

export function listAttachments(params: AttachmentListRequest = {}) {
  const search = new URLSearchParams();
  if (params.purpose) search.set("purpose", params.purpose);
  if (params.status) search.set("status", params.status);
  if (params.search) search.set("search", params.search);
  if (params.reference_status) search.set("reference_status", params.reference_status);
  if (params.category_id) search.set("category_id", String(params.category_id));
  if (params.category_mode) search.set("category_mode", params.category_mode);
  if (params.owner_user_id) search.set("owner_user_id", String(params.owner_user_id));
  if (params.cursor) search.set("cursor", params.cursor);
  if (params.limit) search.set("limit", String(params.limit));
  const query = search.toString();
  return apiFetch<AttachmentListResponse>(`/bff/attachments${query ? `?${query}` : ""}`, { method: "GET" });
}

export function uploadLocalAttachment(formData: FormData) {
  return apiFetch<AttachmentResponse>("/bff/attachments", {
    method: "POST",
    body: formData,
    headers: {}
  });
}

export function presignAttachment(payload: AttachmentPresignRequest) {
  return apiFetch<AttachmentPresignResponse>("/bff/attachments/upload-url", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function completeAttachment(id: number, payload: AttachmentCompleteRequest) {
  return apiFetch<AttachmentResponse>(`/bff/attachments/${id}/complete`, {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function getAttachment(id: number) {
  return apiFetch<AttachmentResponse>(`/bff/attachments/${id}`, { method: "GET" });
}

export function updateAttachment(id: number, payload: AttachmentPatchRequest) {
  return apiFetch<AttachmentResponse>(`/bff/attachments/${id}`, {
    method: "PATCH",
    body: JSON.stringify(payload)
  });
}

export function deleteAttachment(id: number) {
  return apiFetch<DeleteAttachmentResponse>(`/bff/attachments/${id}`, { method: "DELETE" });
}

export function listAttachmentReferences(attachmentId?: number) {
  const search = new URLSearchParams();
  if (attachmentId) search.set("attachment_id", String(attachmentId));
  const query = search.toString();
  return apiFetch<AttachmentReferenceListResponse>(`/bff/attachments/references${query ? `?${query}` : ""}`, { method: "GET" });
}

export function getAttachmentReferences(id: number) {
  return apiFetch<AttachmentReferenceListResponse>(`/bff/attachments/${id}/references`, { method: "GET" });
}

export function replaceAttachmentCategories(id: number, payload: AttachmentCategoryReplaceRequest) {
  return apiFetch<Record<string, never>>(`/bff/attachments/${id}/categories`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export function attachmentContentUrl(id: number) {
  return `/api/bff/attachments/${id}/content`;
}

export function listAttachmentCategories() {
  return apiFetch<AttachmentCategoryListResponse>("/bff/attachment/categories", { method: "GET" });
}

export function createAttachmentCategory(payload: AttachmentCategoryCreateRequest) {
  return apiFetch<AttachmentCategoryResponse>("/bff/attachment/categories", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function updateAttachmentCategory(id: number, payload: AttachmentCategoryPatchRequest) {
  return apiFetch<AttachmentCategoryResponse>(`/bff/attachment/categories/${id}`, {
    method: "PATCH",
    body: JSON.stringify(payload)
  });
}

export function deleteAttachmentCategory(id: number) {
  return apiFetch<DeleteAttachmentCategoryResponse>(`/bff/attachment/categories/${id}`, { method: "DELETE" });
}
