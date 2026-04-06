import request from '../utils/request';

export interface Category {
  id: number;
  name: string;
  slug: string;
  description: string;
  articleCount: number;
  createdAt: string;
}

export interface Tag {
  id: number;
  name: string;
  slug: string;
  color: string;
  articleCount: number;
  createdAt: string;
}

export const getCategories = () => {
  return request.get<any, { code: number; message: string; data: { items: Category[], total: number } }>('/api/v1/admin/categories');
};

export const getTags = () => {
  return request.get<any, { code: number; message: string; data: { items: Tag[], total: number } }>('/api/v1/admin/tags');
};
