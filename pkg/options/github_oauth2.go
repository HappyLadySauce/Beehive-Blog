package options

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/spf13/pflag"

	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

// GithubOAuth2Options holds configuration for the GitHub OAuth2 authorization code flow.
// GithubOAuth2Options 保存 GitHub OAuth2 授权码流程的配置。
type GithubOAuth2Options struct {
	// Enabled toggles GitHub OAuth2 login; when false, endpoints are unreachable and validation is skipped.
	// Enabled 启用 GitHub OAuth2 登录；false 时端点不可达且跳过校验。
	Enabled bool `json:"enabled" mapstructure:"enabled"`
	// ClientID is the OAuth2 application client ID registered with GitHub.
	// ClientID 是在 GitHub 注册的 OAuth2 应用 Client ID。
	ClientID string `json:"client-id" mapstructure:"client-id"`
	// ClientSecret is the OAuth2 application client secret; excluded from JSON dumps.
	// ClientSecret 为 OAuth2 应用密钥；不在 JSON 序列化中输出。
	ClientSecret string `json:"-" mapstructure:"client-secret"`
	// RedirectURL is the callback URL registered with GitHub after user authorization.
	// RedirectURL 为用户授权后 GitHub 回调的注册 URL。
	RedirectURL string `json:"redirect-url" mapstructure:"redirect-url"`
	// AuthURL is the GitHub authorization endpoint.
	// AuthURL 为 GitHub 授权端点。
	AuthURL string `json:"auth-url" mapstructure:"auth-url"`
	// TokenURL is the GitHub token exchange endpoint.
	// TokenURL 为 GitHub 令牌交换端点。
	TokenURL string `json:"token-url" mapstructure:"token-url"`
	// UserInfoURL is the GitHub API endpoint for fetching the authenticated user profile.
	// UserInfoURL 为获取已认证用户信息的 GitHub API 端点。
	UserInfoURL string `json:"user-info-url" mapstructure:"user-info-url"`
	// AllowNonGitHubEndpoints permits non-default OAuth endpoints for local integration tests only.
	// AllowNonGitHubEndpoints 仅允许本地集成测试使用非默认 OAuth 端点。
	AllowNonGitHubEndpoints bool `json:"allow-non-github-endpoints" mapstructure:"allow-non-github-endpoints"`
}

// NewGithubOAuth2Options returns an empty GithubOAuth2Options; defaults are applied via AddFlags.
// NewGithubOAuth2Options 返回空的 GithubOAuth2Options；默认值通过 AddFlags 写入。
func NewGithubOAuth2Options() *GithubOAuth2Options {
	return &GithubOAuth2Options{}
}

// Validate enforces that ClientID, ClientSecret, and RedirectURL are non-empty
// and that all endpoint URLs are valid.
// Validate 强制 ClientID、ClientSecret、RedirectURL 非空，并校验所有端点 URL 格式合法。
func (g *GithubOAuth2Options) Validate() error {
	if !g.Enabled {
		return nil
	}
	var err error
	if g.ClientID == "" {
		err = errors.Join(err, fmt.Errorf("github client-id is required"))
	}
	if g.ClientSecret == "" {
		err = errors.Join(err, fmt.Errorf("github client-secret is required"))
	}
	if g.RedirectURL == "" {
		err = errors.Join(err, fmt.Errorf("github redirect-url is required"))
	}
	if g.AuthURL == "" {
		err = errors.Join(err, fmt.Errorf("github auth-url is required"))
	}
	if g.TokenURL == "" {
		err = errors.Join(err, fmt.Errorf("github token-url is required"))
	}
	if g.UserInfoURL == "" {
		err = errors.Join(err, fmt.Errorf("github user-info-url is required"))
	}
	// Validate URL formats for non-empty values.
	// 对非空值校验 URL 格式。
	for _, endpoint := range []struct {
		name string
		val  string
	}{
		{"redirect-url", g.RedirectURL},
		{"auth-url", g.AuthURL},
		{"token-url", g.TokenURL},
		{"user-info-url", g.UserInfoURL},
	} {
		if endpoint.val != "" {
			parsed, parseErr := parseAbsoluteHTTPURL(endpoint.name, endpoint.val)
			if parseErr != nil {
				err = errors.Join(err, parseErr)
				continue
			}
			if endpoint.name != "redirect-url" && !g.AllowNonGitHubEndpoints {
				if endpointErr := validateGitHubEndpoint(endpoint.name, parsed); endpointErr != nil {
					err = errors.Join(err, endpointErr)
				}
			}
		}
	}
	return err
}

// ToApplicationSettings maps CLI/env/file knobs into the persisted settings shape and validates.
// ToApplicationSettings 将 CLI/环境/文件配置映射为持久化设置形态并校验。
func (g *GithubOAuth2Options) ToApplicationSettings() (settingtypes.ApplicationSettings, error) {
	if g == nil {
		return settingtypes.DefaultApplicationSettings(), nil
	}
	s := settingtypes.ApplicationSettings{
		GithubOAuth2: settingtypes.GithubOAuth2Settings{
			Enabled:                 g.Enabled,
			ClientID:                g.ClientID,
			ClientSecret:            g.ClientSecret,
			RedirectURL:             g.RedirectURL,
			AuthURL:                 g.AuthURL,
			TokenURL:                g.TokenURL,
			UserInfoURL:             g.UserInfoURL,
			AllowNonGitHubEndpoints: g.AllowNonGitHubEndpoints,
		},
	}
	s.Normalize()
	if err := s.Validate(); err != nil {
		return settingtypes.ApplicationSettings{}, err
	}
	return s, nil
}

func parseAbsoluteHTTPURL(name, raw string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("github %s is not a valid URL: %w", name, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("github %s must use http or https scheme", name)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("github %s must be an absolute URL with host", name)
	}
	return parsed, nil
}

func validateGitHubEndpoint(name string, parsed *url.URL) error {
	if parsed.Scheme != "https" {
		return fmt.Errorf("github %s must use https scheme", name)
	}
	expectedHost := "github.com"
	if name == "user-info-url" {
		expectedHost = "api.github.com"
	}
	if parsed.Hostname() != expectedHost {
		return fmt.Errorf("github %s host must be %s", name, expectedHost)
	}
	return nil
}

// AddFlags registers GitHub OAuth2 flags on the supplied FlagSet.
// AddFlags 将 GitHub OAuth2 相关命令行标志注册到给定的 FlagSet。
func (g *GithubOAuth2Options) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&g.Enabled, "github-enabled", g.Enabled, "Enable GitHub OAuth2 login; when false, credential validation is skipped")
	fs.StringVar(&g.ClientID, "github-client-id", "", "GitHub OAuth2 application Client ID (required)")
	fs.StringVar(&g.ClientSecret, "github-client-secret", "", "GitHub OAuth2 application Client Secret (required)")
	fs.StringVar(&g.RedirectURL, "github-redirect-url", "", "OAuth2 redirect URL registered with GitHub (e.g., http://localhost:8080/api/v1/auth/callback)")
	fs.StringVar(&g.AuthURL, "github-auth-url", "https://github.com/login/oauth/authorize", "GitHub authorization endpoint")
	fs.StringVar(&g.TokenURL, "github-token-url", "https://github.com/login/oauth/access_token", "GitHub token exchange endpoint")
	fs.StringVar(&g.UserInfoURL, "github-user-info-url", "https://api.github.com/user", "GitHub user info API endpoint")
	fs.BoolVar(&g.AllowNonGitHubEndpoints, "github-allow-non-github-endpoints", false, "Allow non-GitHub OAuth endpoints for local integration tests only")
}
