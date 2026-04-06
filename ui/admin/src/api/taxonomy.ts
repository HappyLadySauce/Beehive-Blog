import request from '../utils/request';

export interface CategoryBrief {
  id: number;
  name: string;
  slug: string;
  description: string;
  articleCount: number;
  sortOrder: number;
}

export interface AdminCategoryListResponse {
  list: CategoryBrief[];
  total: number;
  page: number;
  pageSize: number;
}

export interface CreateCategoryRequest {
  name: string;
  slug?: string;
  description?: string;
  sortOrder?: number;
}

export interface UpdateCategoryRequest {
  name?: string;
  slug?: string;
  description?: string;
  sortOrder?: number;
}

export interface TagListItem {
  id: number;
  name: string;
  slug: string;
  color: string;
  description: string;
  articleCount: number;
  sortOrder: number;
  createdAt?: string;
}

export interface TagListResponse {
  list: TagListItem[];
  total: number;
  page: number;
  pageSize: number;
}

export interface CreateTagRequest {
  name: string;
  slug?: string;
  color?: string;
  description?: string;
  sortOrder?: number;
}

export interface UpdateTagRequest {
  name?: string;
  slug?: string;
  color?: string;
  description?: string;
  sortOrder?: number;
}

type ApiResponse<T> = { code: number; message: string; data: T };

export const getCategories = (params?: { page?: number; pageSize?: number }) =>
  request.get<any, ApiResponse<AdminCategoryListResponse>>('/api/v1/admin/categories', { params });

export const createCategory = (data: CreateCategoryRequest) =>
  request.post<any, ApiResponse<CategoryBrief>>('/api/v1/admin/categories', data);

export const updateCategory = (id: number, data: UpdateCategoryRequest) =>
  request.put<any, ApiResponse<CategoryBrief>>(`/api/v1/admin/categories/${id}`, data);

export const deleteCategory = (id: number) =>
  request.delete<any, ApiResponse<{ id: number }>>(`/api/v1/admin/categories/${id}`);

export const getTags = (params?: { page?: number; pageSize?: number; keyword?: string; sort?: string }) =>
  request.get<any, ApiResponse<TagListResponse>>('/api/v1/admin/tags', { params });

export const createTag = (data: CreateTagRequest) =>
  request.post<any, ApiResponse<TagListItem>>('/api/v1/admin/tags', data);

export const updateTag = (id: number, data: UpdateTagRequest) =>
  request.put<any, ApiResponse<TagListItem>>(`/api/v1/admin/tags/${id}`, data);

export const deleteTag = (id: number, force = false) =>
  request.delete<any, ApiResponse<{ id: number }>>(`/api/v1/admin/tags/${id}`, { params: { force } });
