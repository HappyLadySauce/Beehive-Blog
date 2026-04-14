import request from '../utils/request';

export interface LoginParams {
  account?: string;
  username?: string;
  email?: string;
  password?: string;
}

export interface LoginResponse {
  token: string;
  refreshToken: string;
  expiresIn: number;
}

export interface UserInfo {
  id: number;
  username: string;
  nickname: string;
  email: string;
  avatar: string;
  role: string;
}

export const login = (data: LoginParams) => {
  return request.post<any, { code: number; message: string; data: LoginResponse }>('/api/v1/auth/login', data);
};

export const getUserInfo = () => {
  return request.get<any, { code: number; message: string; data: UserInfo }>('/api/v1/user/me');
};

export const logout = () => {
  return request.post<any, any>('/api/v1/user/logout');
};
