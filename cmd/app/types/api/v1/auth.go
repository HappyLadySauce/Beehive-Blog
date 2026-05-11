package v1

// AuthToken is an OAuth2-style bearer credential bundle issued by the server.
// Reusable across register / login / refresh responses.
// AuthToken 为服务端签发的 OAuth2 风格 Bearer 凭证集合，可在注册 / 登录 / 刷新响应中复用。
type AuthToken struct {
	// AccessToken is the short-lived bearer token used in the Authorization header.
	// AccessToken 为短期 Bearer 令牌，用于 Authorization 请求头。
	AccessToken string `json:"access_token"`
	// TokenType is the credential scheme; "Bearer" for OAuth2-compatible clients.
	// TokenType 为凭证类型；OAuth2 客户端固定为 "Bearer"。
	TokenType string `json:"token_type"`
	// ExpiresIn is the access token lifetime in seconds.
	// ExpiresIn 为访问令牌的有效期（秒）。
	ExpiresIn int64 `json:"expires_in"`
	// RefreshToken is an optional long-lived credential to mint new access tokens; omit when not issued.
	// RefreshToken 为可选的长期凭证，用于刷新访问令牌；未签发时不输出该字段。
	RefreshToken string `json:"refresh_token,omitempty"`
}

// Grant type constants for LoginRequest.
// LoginRequest 的授权类型常量。
const (
	// GrantTypeLocal is the OAuth2 "local" grant for local username/password login.
	// GrantTypeLocal 为本地用户名/密码登录的 OAuth2 "local" 模式。
	GrantTypeLocal = "local"
	// GrantTypeGitHubOAuth2 is the OAuth2 "github_oauth2" grant for GitHub OAuth2 login.
	// GrantTypeGitHubOAuth2 为 GitHub OAuth2 登录的 OAuth2 "github_oauth2" 模式。
	GrantTypeGitHubOAuth2 = "github_oauth2"
)

type LoginRequest struct {
	// GrantType selects the authentication method: "local" or "github_oauth2".
	// GrantType 选择认证方式："local" 或 "github_oauth2"。
	GrantType string `json:"grant_type" binding:"required"`
	// Account is the local login account; required when grant_type is "local".
	// Account 为本地登录名；grant_type 为 "local" 时必填。
	Account string `json:"account" binding:"omitempty,max=64"`
	// Password is the local plaintext credential; required when grant_type is "local".
	// Password 为本地明文密码；grant_type 为 "local" 时必填。
	Password string `json:"password" binding:"omitempty,max=72"`
	// Code is the OAuth2 authorization code; required when grant_type is "github_oauth2".
	// Code 为 OAuth2 授权码；grant_type 为 "github_oauth2" 时必填。
	Code string `json:"code" binding:"omitempty"`
	// State is the OAuth2 state parameter for CSRF protection.
	// State 是 OAuth2 的 state 参数，用于 CSRF 防护。
	State string `json:"state" binding:"omitempty"`
}

type LoginResponse struct {
	// Token is the auth credential bundle granted on successful login (auto-login).
	// Token 为登录成功后自动签发的鉴权凭证集合（自动登录）。
	Token AuthToken `json:"token"`
}

// GithubOAuthBeginResponse is returned by GET /auth/github/authorize before redirecting to GitHub.
// GithubOAuthBeginResponse 为跳转 GitHub 前 GET /auth/github/authorize 的响应体。
type GithubOAuthBeginResponse struct {
	// State must be echoed to POST /auth/login (github_oauth2) and matches a Redis one-time entry.
	// State 必须在 POST /auth/login（github_oauth2）回传，并与 Redis 中一次性条目匹配。
	State string `json:"state"`
	// AuthURL is the full GitHub authorize URL including client_id, redirect_uri, scope, and state.
	// AuthURL 为完整的 GitHub 授权 URL（含 client_id、redirect_uri、scope、state）。
	AuthURL string `json:"auth_url"`
}

// RefreshRequest carries a refresh JWT to exchange for a new access token.
// RefreshRequest 携带用于换取新 access 的 refresh JWT。
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshResponse returns a rotated access + refresh token bundle.
// RefreshResponse 返回轮换后的 access + refresh 令牌集合。
type RefreshResponse struct {
	Token AuthToken `json:"token"`
}

// AuthSessionResponse is returned after the access token has been verified by AuthMiddleware.
// AuthSessionResponse 在 AuthMiddleware 完成 access token 校验后返回。
type AuthSessionResponse struct {
	// UID is the authenticated user ID from the verified access token.
	// UID 为已验证 access token 中的用户 ID。
	UID int64 `json:"uid"`
	// Role is the authenticated user's authorization role.
	// Role 为已认证用户的授权角色。
	Role string `json:"role"`
	// Exp is the access token expiration time as a Unix timestamp.
	// Exp 为 access token 的过期时间（Unix 时间戳）。
	Exp int64 `json:"exp"`
	// SID is the server-side session ID bound to the access token.
	// SID 为 access token 绑定的服务端会话 ID。
	SID int64 `json:"sid,omitempty"`
}
