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
