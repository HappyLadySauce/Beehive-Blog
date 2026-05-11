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

export type SettingsResponse = {
  revision: number;
  email: EmailSettingsPublic;
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

export type SettingsEmailTestRequest = {
  recipient: string;
};

export type SettingsEmailTestResponse = {
  recipient: string;
};
