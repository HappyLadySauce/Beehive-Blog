import request from '../utils/request';

export interface PageListQuery {
  page?: number;
  pageSize?: number;
  keyword?: string;
  /** comma-separated draft,published,archived,private */
  status?: string;
  sort?: string;
}

export interface AdminPageListItem {
  id: number;
  title: string;
  slug: string;
  status: string;
  viewCount: number;
  isInMenu: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

export interface PageDetailResponse extends AdminPageListItem {
  content: string;
}

export interface AdminPageListResponse {
  list: AdminPageListItem[];
  total: number;
  page: number;
  pageSize: number;
}

export interface CreatePageRequest {
  title: string;
  slug?: string;
  content: string;
  status?: string;
  isInMenu?: boolean;
  sortOrder?: number;
}

export interface UpdatePageRequest {
  title?: string;
  slug?: string;
  content?: string;
  status?: string;
  isInMenu?: boolean;
  sortOrder?: number;
}

export interface UpdatePageStatusRequest {
  status: string;
}

type ApiResponse<T> = { code: number; message: string; data: T };

export const getPages = (params: PageListQuery) =>
  request.get<any, ApiResponse<AdminPageListResponse>>('/api/v1/admin/pages', { params });

export const getPage = (id: number) =>
  request.get<any, ApiResponse<PageDetailResponse>>(`/api/v1/admin/pages/${id}`);

export const createPage = (data: CreatePageRequest) =>
  request.post<any, ApiResponse<PageDetailResponse>>('/api/v1/admin/pages', data);

export const updatePage = (id: number, data: UpdatePageRequest) =>
  request.put<any, ApiResponse<PageDetailResponse>>(`/api/v1/admin/pages/${id}`, data);

export const updatePageStatus = (id: number, data: UpdatePageStatusRequest) =>
  request.put<any, ApiResponse<PageDetailResponse>>(`/api/v1/admin/pages/${id}/status`, data);

export const deletePage = (id: number) =>
  request.delete<any, ApiResponse<{ id: number }>>(`/api/v1/admin/pages/${id}`);

export const getTrashedPages = (params: PageListQuery) =>
  request.get<any, ApiResponse<AdminPageListResponse>>('/api/v1/admin/pages/trash', { params });

export const restorePage = (id: number) =>
  request.post<any, ApiResponse<{ id: number }>>(`/api/v1/admin/pages/${id}/restore`);

export const permanentDeletePage = (id: number) =>
  request.delete<any, ApiResponse<{ id: number }>>(`/api/v1/admin/pages/${id}/permanent`);
