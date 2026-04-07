import request from '../utils/request';

export interface ArticleListQuery {
  page?: number;
  pageSize?: number;
  keyword?: string;
  category?: string;
  tag?: string;
  status?: string;
  author?: string;
  sort?: string;
}

export interface ArticleAuthorItem {
  id: number;
  username: string;
  nickname: string;
  avatar: string;
}

export interface ArticleCategoryItem {
  id: number;
  name: string;
  slug: string;
}

export interface ArticleTagItem {
  id: number;
  name: string;
  slug: string;
  color: string;
}

export interface ArticleListItem {
  id: number;
  title: string;
  slug: string;
  summary: string;
  coverImage: string;
  isPinned: boolean;
  pinOrder: number;
  viewCount: number;
  likeCount: number;
  commentCount: number;
  publishedAt: string;
  updatedAt: string;
  author: ArticleAuthorItem;
  category?: ArticleCategoryItem;
  tags: ArticleTagItem[];
}

export interface AdminArticleListItem extends ArticleListItem {
  status: string;
}

export interface ArticleDetailResponse extends AdminArticleListItem {
  content: string;
  protected: boolean;
}

export interface AdminArticleListResponse {
  list: AdminArticleListItem[];
  total: number;
  page: number;
  pageSize: number;
}

export interface CreateArticleRequest {
  title: string;
  slug?: string;
  content: string;
  summary?: string;
  coverImage?: string;
  categoryId?: number | null;
  tagIds?: number[];
  status?: string;
  publishedAt?: string;
}

export interface UpdateArticleRequest {
  title?: string;
  slug?: string;
  content?: string;
  summary?: string;
  coverImage?: string;
  categoryId?: number | null;
  tagIds?: number[];
  status?: string;
  publishedAt?: string;
}

export interface BatchArticleRequest {
  action: 'delete' | 'set_status' | 'set_category' | 'set_tags';
  ids: number[];
  payload?: {
    status?: string;
    categoryId?: number;
    tagIds?: number[];
  };
}

type ApiResponse<T> = { code: number; message: string; data: T };

export const getArticles = (params: ArticleListQuery) =>
  request.get<any, ApiResponse<AdminArticleListResponse>>('/api/v1/admin/articles', { params });

export const getArticle = (id: number) =>
  request.get<any, ApiResponse<ArticleDetailResponse>>(`/api/v1/admin/articles/${id}`);

export const createArticle = (data: CreateArticleRequest) =>
  request.post<any, ApiResponse<ArticleDetailResponse>>('/api/v1/admin/articles', data);

export const updateArticle = (id: number, data: UpdateArticleRequest) =>
  request.put<any, ApiResponse<ArticleDetailResponse>>(`/api/v1/admin/articles/${id}`, data);

export const deleteArticle = (id: number) =>
  request.delete<any, ApiResponse<{ id: number }>>(`/api/v1/admin/articles/${id}`);

export const batchOperateArticles = (data: BatchArticleRequest) =>
  request.post<any, ApiResponse<{ affected: number }>>('/api/v1/admin/articles/batch', data);

/** 回收站：仅已软删文章，参数与管理员列表一致（keyword、sort、分页） */
export const getTrashedArticles = (params: ArticleListQuery) =>
  request.get<any, ApiResponse<AdminArticleListResponse>>('/api/v1/admin/articles/trash', { params });

export const restoreArticle = (id: number) =>
  request.post<any, ApiResponse<{ id: number }>>(`/api/v1/admin/articles/${id}/restore`);

export const permanentDeleteArticle = (id: number) =>
  request.delete<any, ApiResponse<{ id: number }>>(`/api/v1/admin/articles/${id}/permanent`);
