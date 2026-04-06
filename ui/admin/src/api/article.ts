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

export interface ArticleItem {
  id: number;
  title: string;
  slug: string;
  summary: string;
  coverImage: string;
  status: string;
  password?: string;
  isPinned: boolean;
  pinOrder: number;
  viewCount: number;
  likeCount: number;
  commentCount: number;
  publishedAt: string;
  createdAt: string;
  updatedAt: string;
  author: {
    id: number;
    username: string;
    nickname: string;
    avatar: string;
  };
  category?: {
    id: number;
    name: string;
    slug: string;
  };
  tags?: Array<{
    id: number;
    name: string;
    slug: string;
    color: string;
  }>;
}

export interface ArticleListResponse {
  items: ArticleItem[];
  total: number;
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

export const getArticles = (params: ArticleListQuery) => {
  return request.get<any, { code: number; message: string; data: ArticleListResponse }>('/api/v1/admin/articles', { params });
};

export const batchOperateArticles = (data: BatchArticleRequest) => {
  return request.post<any, { code: number; message: string; data: { affected: number } }>('/api/v1/admin/articles/batch', data);
};
