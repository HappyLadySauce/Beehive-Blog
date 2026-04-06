import request from '../utils/request';

export interface SettingsResponse {
  group: string;
  settings: Record<string, string>;
}

export const getSettings = (group: string) => {
  return request.get<any, { code: number; message: string; data: SettingsResponse }>(`/api/v1/admin/settings/${group}`);
};

export const updateSettings = (group: string, settings: Record<string, string>) => {
  return request.put<any, { code: number; message: string; data: SettingsResponse }>(`/api/v1/admin/settings/${group}`, { settings });
};

export const testSmtp = (to: string) => {
  return request.post<any, { code: number; message: string }>('/api/v1/admin/settings/smtp/test', { to });
};

export const syncHexoPosts = (rebuild: boolean) => {
  return request.post<any, { code: number; message: string; data: any }>('/api/v1/admin/sync/posts', { rebuild });
};

export const getHexoSyncStatus = () => {
  return request.get<any, { code: number; message: string; data: any }>('/api/v1/admin/sync/status');
};
