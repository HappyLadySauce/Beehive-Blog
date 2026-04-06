import request from '../utils/request';

export interface Attachment {
  id: number;
  name: string;
  originalName: string;
  url: string;
  thumbUrl: string;
  type: string;
  mimeType: string;
  size: number;
  width: number;
  height: number;
  createdAt: string;
}

export const getAttachments = (params?: any) => {
  return request.get<any, { code: number; message: string; data: { items: Attachment[], total: number } }>('/api/v1/admin/attachments', { params });
};

export const deleteAttachment = (id: number) => {
  return request.delete<any, any>(`/api/v1/admin/attachments/${id}`);
};

export const uploadAttachment = (file: File) => {
  const formData = new FormData();
  formData.append('file', file);
  return request.post<any, { code: number; message: string; data: Attachment }>('/api/v1/admin/attachments/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
};
