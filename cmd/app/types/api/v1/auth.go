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