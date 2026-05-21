import { apiFetch } from "./client";
import type {
  CreateCategoryRequest,
  CreateCategoryResponse,
  DeleteCategoryResponse,
  ListCategoriesRequest,
  ListCategoriesResponse,
  CategoryDetailResponse,
  UpdateCategoryRequest
} from "./types";

export function listCategories(params: ListCategoriesRequest = {}) {
  const search = new URLSearchParams();
  if (params.page) search.set("page", String(params.page));
  if (params.page_size) search.set("page_size", String(params.page_size));
  if (params.search) search.set("search", params.search);
  const query = search.toString();
  return apiFetch<ListCategoriesResponse>(`/bff/categories${query ? `?${query}` : ""}`, { method: "GET" });
}

export function getCategory(id: number) {
  return apiFetch<CategoryDetailResponse>(`/bff/categories/${id}`, { method: "GET" });
}

export function createCategory(payload: CreateCategoryRequest) {
  return apiFetch<CreateCategoryResponse>("/bff/categories", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function updateCategory(id: number, payload: UpdateCategoryRequest) {
  return apiFetch<CategoryDetailResponse>(`/bff/categories/${id}`, {
    method: "PATCH",
    body: JSON.stringify(payload)
  });
}

export function deleteCategory(id: number) {
  return apiFetch<DeleteCategoryResponse>(`/bff/categories/${id}`, { method: "DELETE" });
}
