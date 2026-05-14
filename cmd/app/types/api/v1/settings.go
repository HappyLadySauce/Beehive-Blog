package v1

// EmailSettingsPublic is the admin-visible email configuration without secrets.
// EmailSettingsPublic 为管理员可见的邮件配置（不含密钥）。
type EmailSettingsPublic struct {
	Enabled     bool   `json:"enabled"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	PasswordSet bool   `json:"password_set"`
	From        string `json:"from"`
	FromName    string `json:"from_name"`
	TLS         string `json:"tls"`
}

// GithubOAuth2SettingsPublic is the admin-visible GitHub OAuth2 configuration without secrets.
// GithubOAuth2SettingsPublic 为管理员可见的 GitHub OAuth2 配置（不含密钥）。
type GithubOAuth2SettingsPublic struct {
	Enabled                 bool   `json:"enabled"`
	ClientID                string `json:"client_id"`
	ClientSecretSet         bool   `json:"client_secret_set"`
	RedirectURL             string `json:"redirect_url"`
	AuthURL                 string `json:"auth_url"`
	TokenURL                string `json:"token_url"`
	UserInfoURL             string `json:"user_info_url"`
	AllowNonGitHubEndpoints bool   `json:"allow_non_github_endpoints"`
}

// AttachmentRemoteSettingsPublic is the admin-visible remote storage configuration.
// AttachmentRemoteSettingsPublic 为管理员可见的远端存储配置。
type AttachmentRemoteSettingsPublic struct {
	Bucket          string `json:"bucket"`
	UploadBaseURL   string `json:"upload_base_url"`
	DownloadBaseURL string `json:"download_base_url"`
}

// AttachmentSettingsPublic is the admin-visible attachment configuration.
// AttachmentSettingsPublic 为管理员可见的附件配置。
type AttachmentSettingsPublic struct {
	DefaultStorage      string                          `json:"default_storage"`
	LocalRoot           string                          `json:"local_root"`
	MaxBytes            int64                           `json:"max_bytes"`
	AllowedMIMEPrefixes []string                        `json:"allowed_mime_prefixes"`
	PresignTTLSeconds   int64                           `json:"presign_ttl_seconds"`
	S3                  AttachmentRemoteSettingsPublic  `json:"s3"`
	OSS                 AttachmentRemoteSettingsPublic  `json:"oss"`
}

// SettingsResponse is returned by GET /api/v1/settings (sanitized).
// SettingsResponse 为 GET /api/v1/settings 的脱敏响应。
type SettingsResponse struct {
	Revision     int64                      `json:"revision"`
	Email        EmailSettingsPublic        `json:"email"`
	GithubOAuth2 GithubOAuth2SettingsPublic `json:"github_oauth2"`
	Attachment   AttachmentSettingsPublic   `json:"attachment"`
}

// EmailSMTPPatchJSON is the JSON body fragment for PATCH /api/v1/settings (partial email update).
// EmailSMTPPatchJSON 为 PATCH /api/v1/settings 的 email 片段（部分更新）。
type EmailSMTPPatchJSON struct {
	Enabled  *bool   `json:"enabled"`
	Host     *string `json:"host"`
	Port     *int    `json:"port"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	From     *string `json:"from"`
	FromName *string `json:"from_name"`
	TLS      *string `json:"tls"`
}

// GithubOAuth2PatchJSON is the JSON body fragment for PATCH /api/v1/settings (partial GitHub OAuth2 update).
// GithubOAuth2PatchJSON 为 PATCH /api/v1/settings 的 github_oauth2 片段（部分更新）。
type GithubOAuth2PatchJSON struct {
	Enabled                 *bool   `json:"enabled"`
	ClientID                *string `json:"client_id"`
	ClientSecret            *string `json:"client_secret"`
	RedirectURL             *string `json:"redirect_url"`
	AuthURL                 *string `json:"auth_url"`
	TokenURL                *string `json:"token_url"`
	UserInfoURL             *string `json:"user_info_url"`
	AllowNonGitHubEndpoints *bool   `json:"allow_non_github_endpoints"`
}

// AttachmentRemotePatchJSON is the JSON body fragment for remote storage partial update.
// AttachmentRemotePatchJSON 为远端存储部分更新的 JSON 片段。
type AttachmentRemotePatchJSON struct {
	Bucket          *string `json:"bucket"`
	UploadBaseURL   *string `json:"upload_base_url"`
	DownloadBaseURL *string `json:"download_base_url"`
}

// AttachmentPatchJSON is the JSON body fragment for PATCH /api/v1/settings (partial attachment update).
// AttachmentPatchJSON 为 PATCH /api/v1/settings 的 attachment 片段（部分更新）。
type AttachmentPatchJSON struct {
	DefaultStorage      *string                     `json:"default_storage"`
	LocalRoot           *string                     `json:"local_root"`
	MaxBytes            *int64                      `json:"max_bytes"`
	AllowedMIMEPrefixes *[]string                   `json:"allowed_mime_prefixes"`
	PresignTTLSeconds   *int64                      `json:"presign_ttl_seconds"`
	S3                  *AttachmentRemotePatchJSON  `json:"s3"`
	OSS                 *AttachmentRemotePatchJSON  `json:"oss"`
}

// SettingsPatchRequestJSON is the PATCH body; only keys present are merged server-side.
// SettingsPatchRequestJSON 为 PATCH 请求体；仅出现的键在服务端参与合并。
type SettingsPatchRequestJSON struct {
	Email        *EmailSMTPPatchJSON        `json:"email"`
	GithubOAuth2 *GithubOAuth2PatchJSON     `json:"github_oauth2"`
	Attachment   *AttachmentPatchJSON       `json:"attachment"`
}

// SettingsEmailTestRequest is the body for sending a test email with saved SMTP settings.
// SettingsEmailTestRequest 为使用已保存 SMTP 设置发送测试邮件的请求体。
type SettingsEmailTestRequest struct {
	Recipient string `json:"recipient" binding:"required,email,max=320"`
}

// SettingsEmailTestResponse confirms the accepted test recipient.
// SettingsEmailTestResponse 确认测试邮件收件人。
type SettingsEmailTestResponse struct {
	Recipient string `json:"recipient"`
}
