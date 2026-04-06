import request from '../utils/request';

export interface Comment {
  id: number;
  articleId: number;
  articleTitle: string;
  author: {
    id: number;
    nickname: string;
    username: string;
  };
  content: string;
  status: string;
  createdAt: string;
}

export const getComments = (params?: any) => {
  return request.get<any, { code: number; message: string; data: { items: Comment[], total: number } }>('/api/v1/admin/comments', { params });
};

export const updateCommentStatus = (id: number, status: string) => {
  return request.put<any, any>(`/api/v1/admin/comments/${id}/status`, { status });
};
