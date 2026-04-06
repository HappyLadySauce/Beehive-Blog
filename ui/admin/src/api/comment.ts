import request from '../utils/request';

export interface CommentAuthorItem {
  id?: number;
  username?: string;
  nickname?: string;
  avatar?: string;
}

export interface AdminCommentItem {
  id: number;
  content: string;
  status: string;
  articleId: number;
  userId?: number;
  parentId?: number;
  createdAt: string;
  updatedAt: string;
  author: CommentAuthorItem;
}

export interface AdminCommentListQuery {
  page?: number;
  pageSize?: number;
  articleId?: number;
  status?: string;
  keyword?: string;
}

export interface AdminCommentListResponse {
  items: AdminCommentItem[];
  total: number;
  page: number;
  pageSize: number;
}

type ApiResponse<T> = { code: number; message: string; data: T };

export const getComments = (params?: AdminCommentListQuery) =>
  request.get<any, ApiResponse<AdminCommentListResponse>>('/api/v1/admin/comments', { params });

export const updateCommentStatus = (id: number, status: string) =>
  request.put<any, ApiResponse<{ id: number; status: string }>>(`/api/v1/admin/comments/${id}/status`, { status });
