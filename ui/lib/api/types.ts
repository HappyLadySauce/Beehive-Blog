export type BaseResponse<T> = {
  code: number;
  message: string;
  data: T;
};

export type AuthToken = {
  access_token: string;
  token_type: "Bearer" | string;
  expires_in: number;
  refresh_token?: string;
};

export type AuthPayload = {
  token: AuthToken;
};

export type AuthSessionResponse = {
  uid: number;
  role: string;
  exp: number;
  sid?: number;
};

export type GithubOAuthBeginResponse = {
  state: string;
  auth_url: string;
};

export type RegisterRequest = {
  username: string;
  password: string;
  email?: string;
  nickname?: string;
  phone?: string;
};

export type LoginRequest =
  | {
      grant_type: "local";
      account: string;
      password: string;
    }
  | {
      grant_type: "github_oauth2";
      code: string;
      state: string;
    };

export type PublicPost = {
  slug: string;
  title: string;
  description: string;
  body: string;
  publishedAt: string;
  tags: string[];
  readingMinutes: number;
};

export type EmailSettingsPublic = {
  enabled: boolean;
  host: string;
  port: number;
  username: string;
  password_set: boolean;
  from: string;
  from_name: string;
  tls: "none" | "starttls" | "tls" | string;
};

export type GithubOAuth2SettingsPublic = {
  enabled: boolean;
  client_id: string;
  client_secret_set: boolean;
  redirect_url: string;
  auth_url: string;
  token_url: string;
  user_info_url: string;
  allow_non_github_endpoints: boolean;
};

export type AttachmentRemoteSettingsPublic = {
  bucket: string;
  upload_base_url: string;
  download_base_url: string;
};

export type AttachmentSettingsPublic = {
  default_storage: "local" | "s3" | "oss" | string;
  local_root: string;
  max_bytes: number;
  allowed_mime_prefixes: string[];
  presign_ttl_seconds: number;
  s3: AttachmentRemoteSettingsPublic;
  oss: AttachmentRemoteSettingsPublic;
};

export type SettingsResponse = {
  revision: number;
  email: EmailSettingsPublic;
  github_oauth2: GithubOAuth2SettingsPublic;
  attachment: AttachmentSettingsPublic;
};

export type EmailSMTPPatch = {
  enabled?: boolean;
  host?: string;
  port?: number;
  username?: string;
  password?: string;
  from?: string;
  from_name?: string;
  tls?: "none" | "starttls" | "tls" | string;
};

export type SettingsPatchRequest = {
  email: EmailSMTPPatch;
};

export type GithubOAuth2Patch = {
  enabled?: boolean;
  client_id?: string;
  client_secret?: string;
  redirect_url?: string;
  auth_url?: string;
  token_url?: string;
  user_info_url?: string;
  allow_non_github_endpoints?: boolean;
};

export type AttachmentRemotePatch = {
  bucket?: string;
  upload_base_url?: string;
  download_base_url?: string;
};

export type AttachmentPatch = {
  default_storage?: string;
  local_root?: string;
  max_bytes?: number;
  allowed_mime_prefixes?: string[];
  presign_ttl_seconds?: number;
  s3?: AttachmentRemotePatch;
  oss?: AttachmentRemotePatch;
};

export type SettingsEmailTestRequest = {
  recipient: string;
};

export type SettingsEmailTestResponse = {
  recipient: string;
};
