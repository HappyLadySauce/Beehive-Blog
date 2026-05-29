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

// ProfileSettingsPublic is the public author profile shown on reader-facing pages.
// ProfileSettingsPublic 是读者侧页面展示的公开博主资料。
type ProfileSettingsPublic struct {
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Headline    string `json:"headline"`
	Bio         string `json:"bio"`
	Location    string `json:"location"`
	Website     string `json:"website"`
}

// SiteSettingsPublic is the public site identity configuration without secrets.
// SiteSettingsPublic 是不含敏感信息的公开站点身份配置。
type SiteSettingsPublic struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Subtitle    string `json:"subtitle"`
	Description string `json:"description"`
	Keywords    string `json:"keywords"`
	LogoURL     string `json:"logo_url"`
	FaviconURL  string `json:"favicon_url"`
	ICPBeian    string `json:"icp_beian"`
	PoliceBeian string `json:"police_beian"`
	FooterText  string `json:"footer_text"`
}

// SettingsResponse is returned by GET /api/v1/settings (sanitized).
// SettingsResponse 为 GET /api/v1/settings 的脱敏响应。
type SettingsResponse struct {
	Revision     int64                      `json:"revision"`
	Email        EmailSettingsPublic        `json:"email"`
	GithubOAuth2 GithubOAuth2SettingsPublic `json:"github_oauth2"`
	Profile      ProfileSettingsPublic      `json:"profile"`
	Site         SiteSettingsPublic         `json:"site"`
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

// ProfilePatchJSON is the JSON body fragment for PATCH /api/v1/settings/profile.
// ProfilePatchJSON 为 PATCH /api/v1/settings/profile 的 JSON 片段。
type ProfilePatchJSON struct {
	DisplayName *string `json:"display_name" binding:"omitempty,max=80"`
	AvatarURL   *string `json:"avatar_url" binding:"omitempty,max=512"`
	Headline    *string `json:"headline" binding:"omitempty,max=160"`
	Bio         *string `json:"bio" binding:"omitempty,max=1000"`
	Location    *string `json:"location" binding:"omitempty,max=120"`
	Website     *string `json:"website" binding:"omitempty,max=512"`
}

// SitePatchJSON is the JSON body fragment for PATCH /api/v1/settings/site.
// SitePatchJSON 为 PATCH /api/v1/settings/site 的 JSON 片段。
type SitePatchJSON struct {
	Name        *string `json:"name" binding:"omitempty,max=80"`
	URL         *string `json:"url" binding:"omitempty,max=512"`
	Subtitle    *string `json:"subtitle" binding:"omitempty,max=160"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	Keywords    *string `json:"keywords" binding:"omitempty,max=512"`
	LogoURL     *string `json:"logo_url" binding:"omitempty,max=512"`
	FaviconURL  *string `json:"favicon_url" binding:"omitempty,max=512"`
	ICPBeian    *string `json:"icp_beian" binding:"omitempty,max=120"`
	PoliceBeian *string `json:"police_beian" binding:"omitempty,max=120"`
	FooterText  *string `json:"footer_text" binding:"omitempty,max=512"`
}

// SettingsPatchRequestJSON is the PATCH body; only keys present are merged server-side.
// SettingsPatchRequestJSON 为 PATCH 请求体；仅出现的键在服务端参与合并。
type SettingsPatchRequestJSON struct {
	Email        *EmailSMTPPatchJSON    `json:"email"`
	GithubOAuth2 *GithubOAuth2PatchJSON `json:"github_oauth2"`
	Profile      *ProfilePatchJSON      `json:"profile"`
	Site         *SitePatchJSON         `json:"site"`
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
