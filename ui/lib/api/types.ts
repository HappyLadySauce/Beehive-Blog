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

export type SettingsResponse = {
  revision: number;
  email: EmailSettingsPublic;
  github_oauth2: GithubOAuth2SettingsPublic;
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

export type SettingsEmailTestRequest = {
  recipient: string;
};

export type SettingsEmailTestResponse = {
  recipient: string;
};

export type UserItem = {
  id: number;
  username: string;
  email: string | null;
  nickname: string | null;
  phone: string | null;
  avatar_attachment_id: number | null;
  role: string;
  status: string;
  last_login_at: string | null;
  created_at: string;
  updated_at: string;
};

export type ListUsersRequest = {
  page?: number;
  page_size?: number;
  status?: string;
  role?: string;
  search?: string;
};

export type ListUsersResponse = {
  items: UserItem[];
  total: number;
  page: number;
  page_size: number;
};

export type UserDetailResponse = UserItem;

export type CreateUserRequest = {
  username: string;
  password?: string;
  email?: string;
  nickname?: string;
  phone?: string;
  role?: string;
  status?: string;
};

export type CreateUserResponse = {
  id: number;
};

export type UpdateUserRequest = {
  username?: string;
  email?: string | null;
  nickname?: string | null;
  phone?: string | null;
  role?: string;
  status?: string;
  password?: string;
  avatar_attachment_id?: number;
};

export type DeleteUserResponse = Record<string, never>;

export type JsonObject = Record<string, unknown>;

export type DriverResponse = {
  id: number;
  name: string;
  display_name: string;
  description?: string | null;
  config_schema: JsonObject;
  capabilities: JsonObject;
  status: string;
  created_at: string;
  updated_at: string;
};

export type DriverListResponse = {
  items: DriverResponse[];
};

export type StorageMountResponse = {
  id: number;
  driver_name: string;
  mount_path: string;
  name: string;
  remark?: string | null;
  config: JsonObject;
  order_index: number;
  is_default: boolean;
  disabled: boolean;
  status: string;
  last_checked_at?: string | null;
  last_error?: string | null;
  created_by?: number | null;
  created_at: string;
  updated_at: string;
};

export type StorageMountListResponse = {
  items: StorageMountResponse[];
};

export type StorageMountCreateRequest = {
  driver_name: string;
  mount_path: string;
  name: string;
  remark?: string | null;
  config: JsonObject;
  order_index?: number;
  is_default?: boolean;
};

export type StorageMountPatchRequest = {
  name?: string;
  remark?: string | null;
  config?: JsonObject;
  order_index?: number;
  is_default?: boolean;
};

export type StorageMountCheckResponse = {
  status: string;
  error?: string | null;
  checked: string;
};

export type DeleteStorageMountResponse = Record<string, never>;

export type AttachmentResponse = {
  id: number;
  owner_user_id?: number | null;
  purpose: string;
  filename: string;
  original_name?: string | null;
  mime_type: string;
  size: number;
  storage_mount_id: number;
  file_node_id?: number | null;
  object_key: string;
  storage_metadata?: JsonObject | null;
  etag?: string | null;
  checksum?: string | null;
  access_scope: "private" | "public" | string;
  upload_status: "pending" | "ready" | string;
  status: "active" | "archived" | string;
  category_ids?: number[];
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
};

export type AttachmentListRequest = {
  purpose?: string;
  status?: string;
  search?: string;
  reference_status?: "referenced" | "orphan" | string;
  category_id?: number;
  category_mode?: "unassigned" | string;
  owner_user_id?: number;
  cursor?: string;
  limit?: number;
  page?: number;
  page_size?: number;
};

export type AttachmentListResponse = {
  items: AttachmentResponse[];
  next_cursor?: string;
  total?: number;
  page?: number;
  page_size?: number;
};

export type AttachmentReferenceResponse = {
  attachment_id: number;
  source_type: string;
  source_id: number;
  source_title: string;
  relation: string;
  status: string;
  updated_at: string;
};

export type AttachmentReferenceListResponse = {
  items: AttachmentReferenceResponse[];
};

export type AttachmentPresignRequest = {
  owner_user_id?: number;
  purpose: string;
  filename: string;
  original_name?: string | null;
  mime_type: string;
  size: number;
  access_scope: "private" | "public" | string;
  checksum?: string | null;
  category_ids?: number[];
  storage_mount_id?: number;
};

export type AttachmentPresignResponse = {
  attachment: AttachmentResponse;
  upload_url: string;
  method: string;
  headers?: Record<string, string>;
  expires_at: string;
};

export type AttachmentCompleteRequest = {
  etag?: string | null;
  checksum?: string | null;
  size?: number;
};

export type AttachmentPatchRequest = {
  original_name?: string | null;
  status?: string;
  access_scope?: string;
  category_ids?: number[];
};

export type AttachmentCategoryReplaceRequest = {
  category_ids: number[];
};

export type AttachmentCategoryResponse = {
  id: number;
  parent_id?: number | null;
  name: string;
  slug: string;
  description?: string | null;
  icon?: string | null;
  path: string;
  depth: number;
  sort_order: number;
  status: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
};

export type AttachmentCategoryListResponse = {
  items: AttachmentCategoryResponse[];
};

export type AttachmentCategoryCreateRequest = {
  parent_id?: number | null;
  name: string;
  slug: string;
  description?: string | null;
  icon?: string | null;
  sort_order?: number;
  status?: string;
};

export type AttachmentCategoryPatchRequest = {
  parent_id?: number | null;
  name?: string;
  slug?: string;
  description?: string | null;
  icon?: string | null;
  sort_order?: number;
  status?: string;
};

export type DeleteAttachmentResponse = Record<string, never>;
export type DeleteAttachmentCategoryResponse = Record<string, never>;
