import request from '../utils/request';

export interface Attachment {
  id: number;
  name: string;
  originalName: string;
  url: string;
  thumbUrl?: string;
  type: string;
  mimeType: string;
  size: number;
  width?: number;
  height?: number;
  parentId?: number | null;
  variant?: string;
  groupId?: number | null;
  groupName?: string;
  refArticleCount?: number;
  createdAt: string;
}

export interface AttachmentListParams {
  page?: number;
  pageSize?: number;
  keyword?: string;
  type?: string;
  groupId?: number;
  rootsOnly?: boolean;
}

export interface AttachmentListResponse {
  items: Attachment[];
  total: number;
  page: number;
  pageSize: number;
}

export interface AttachmentFamilyArticleRef {
  articleId: number;
  title: string;
  slug: string;
}

export interface AttachmentMemberArticleRefs {
  attachmentId: number;
  articles: AttachmentFamilyArticleRef[];
}

export interface AttachmentFamilyResponse {
  root: Attachment;
  children: Attachment[];
  memberReferences: AttachmentMemberArticleRefs[];
}

export interface AttachmentProcessingSettings {
  defaultQuality: number;
  allowedFormats: string[];
}

export type ProcessOutputFormat = 'jpeg' | 'png' | 'gif' | 'webp' | 'ico';

export interface ProcessAttachmentRequest {
  quality?: number;
  /** 省略或与源相同：不传 format 字段 */
  format?: ProcessOutputFormat | '';
}

export interface ReplaceAttachmentInArticlesRequest {
  fromAttachmentId: number;
  toAttachmentId: number;
  articleIds: number[];
}

export interface ReplaceAttachmentInArticlesResponse {
  updated: number;
}

export interface AttachmentGroupItem {
  id: number;
  name: string;
  parentId?: number | null;
  sortOrder: number;
}

export const getAttachments = (params?: AttachmentListParams) => {
  return request.get<any, { code: number; message: string; data: AttachmentListResponse }>(
    '/api/v1/admin/attachments',
    { params },
  );
};

export const deleteAttachment = (id: number) => {
  return request.delete<any, any>(`/api/v1/admin/attachments/${id}`);
};

export const uploadAttachment = (file: File, groupId?: number) => {
  const formData = new FormData();
  formData.append('file', file);
  const params = groupId != null && groupId > 0 ? { groupId } : {};
  return request.post<any, { code: number; message: string; data: Attachment }>(
    '/api/v1/admin/attachments/upload',
    formData,
    {
      params,
      headers: { 'Content-Type': 'multipart/form-data' },
    },
  );
};

export const getAttachmentFamily = (id: number) => {
  return request.get<any, { code: number; message: string; data: AttachmentFamilyResponse }>(
    `/api/v1/admin/attachments/${id}/family`,
  );
};

export const processAttachment = (id: number, body: ProcessAttachmentRequest) => {
  const payload: Record<string, unknown> = {};
  if (body.quality != null) payload.quality = body.quality;
  if (body.format) payload.format = body.format;
  return request.post<any, { code: number; message: string; data: Attachment }>(
    `/api/v1/admin/attachments/${id}/process`,
    payload,
  );
};

/** 至少提供 name 或 groupId 之一。groupId 为 0 表示清除分类。 */
export interface UpdateAttachmentRequest {
  name?: string;
  groupId?: number;
}

export const updateAttachment = (id: number, body: UpdateAttachmentRequest) => {
  return request.put<any, { code: number; message: string; data: Attachment }>(
    `/api/v1/admin/attachments/${id}`,
    body,
  );
};

export const replaceAttachmentInArticles = (body: ReplaceAttachmentInArticlesRequest) => {
  return request.post<any, { code: number; message: string; data: ReplaceAttachmentInArticlesResponse }>(
    '/api/v1/admin/attachments/replace-in-articles',
    body,
  );
};

export const getAttachmentProcessingSettings = () => {
  return request.get<any, { code: number; message: string; data: AttachmentProcessingSettings }>(
    '/api/v1/admin/attachments/settings',
  );
};

export const putAttachmentProcessingSettings = (body: AttachmentProcessingSettings) => {
  return request.put<any, { code: number; message: string; data: AttachmentProcessingSettings }>(
    '/api/v1/admin/attachments/settings',
    body,
  );
};

export const listAttachmentGroups = () => {
  return request.get<any, { code: number; message: string; data: AttachmentGroupItem[] }>(
    '/api/v1/admin/attachment-groups',
  );
};

export const createAttachmentGroup = (body: { name: string; parentId?: number | null; sortOrder?: number }) => {
  return request.post<any, { code: number; message: string; data: AttachmentGroupItem }>(
    '/api/v1/admin/attachment-groups',
    body,
  );
};

export const deleteAttachmentGroup = (id: number) => {
  return request.delete<any, { code: number; message: string; data: { id: number } }>(
    `/api/v1/admin/attachment-groups/${id}`,
  );
};
