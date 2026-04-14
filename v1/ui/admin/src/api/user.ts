import request from '../utils/request';

export type UserRole = 'guest' | 'user' | 'admin';
export type UserStatus = 'active' | 'inactive' | 'disabled' | 'deleted';

export interface AdminUserItem {
  id: number;
  username: string;
  nickname: string;
  email: string;
  avatar: string;
  role: UserRole;
  status: UserStatus;
  lastLoginAt?: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface AdminUserListQuery {
  page?: number;
  pageSize?: number;
  keyword?: string;
  role?: UserRole;
  status?: UserStatus;
}

export interface AdminUserListResponse {
  items: AdminUserItem[];
  total: number;
  page: number;
  pageSize: number;
}

export interface AdminCreateUserRequest {
  username: string;
  password: string;
  nickname?: string;
  email: string;
  avatar?: string;
  role?: UserRole;
  status?: Exclude<UserStatus, 'deleted'>;
}

export interface AdminUpdateUserRequest {
  nickname?: string;
  email?: string;
  avatar?: string;
  role?: UserRole;
}

export interface AdminUpdateUserStatusRequest {
  status: Exclude<UserStatus, 'deleted'>;
}

export interface AdminResetUserPasswordRequest {
  newPassword: string;
}

type ApiResponse<T> = { code: number; message: string; data: T };

export const getAdminUsers = (params?: AdminUserListQuery) =>
  request.get<any, ApiResponse<AdminUserListResponse>>('/api/v1/admin/users', { params });

export const getAdminUserDetail = (id: number) =>
  request.get<any, ApiResponse<AdminUserItem>>(`/api/v1/admin/users/${id}`);

export const createAdminUser = (data: AdminCreateUserRequest) =>
  request.post<any, ApiResponse<AdminUserItem>>('/api/v1/admin/users', data);

export const updateAdminUser = (id: number, data: AdminUpdateUserRequest) =>
  request.put<any, ApiResponse<AdminUserItem>>(`/api/v1/admin/users/${id}`, data);

export const updateAdminUserStatus = (id: number, data: AdminUpdateUserStatusRequest) =>
  request.put<any, ApiResponse<AdminUserItem>>(`/api/v1/admin/users/${id}/status`, data);

export const resetAdminUserPassword = (id: number, data: AdminResetUserPasswordRequest) =>
  request.put<any, ApiResponse<{ id: number; message: string }>>(`/api/v1/admin/users/${id}/password/reset`, data);

export const deleteAdminUser = (id: number) =>
  request.delete<any, ApiResponse<{ id: number }>>(`/api/v1/admin/users/${id}`);
