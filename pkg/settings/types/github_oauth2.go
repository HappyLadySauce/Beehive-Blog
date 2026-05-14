package types

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Default endpoint URLs for GitHub OAuth2.
// GitHub OAuth2 默认端点 URL。
const (
	DefaultGitHubAuthURL     = "https://github.com/login/oauth/authorize"
	DefaultGitHubTokenURL    = "https://github.com/login/oauth/access_token"
	DefaultGitHubUserInfoURL = "https://api.github.com/user"
)

// GithubOAuth2Settings holds GitHub OAuth2 configuration persisted in the database.
// GithubOAuth2Settings 保存于数据库的 GitHub OAuth2 配置。
type GithubOAuth2Settings struct {
	Enabled                 bool   `json:"enabled"`
	ClientID                string `json:"client_id"`
	ClientSecret            string `json:"client_secret"`
	RedirectURL             string `json:"redirect_url"`
	AuthURL                 string `json:"auth_url"`
	TokenURL                string `json:"token_url"`
	UserInfoURL             string `json:"user_info_url"`
	AllowNonGitHubEndpoints bool   `json:"allow_non_github_endpoints"`
}

// GithubOAuth2Patch uses pointers so omitted JSON fields leave existing values unchanged.
// GithubOAuth2Patch 使用指针，JSON 省略的字段保留原值。
type GithubOAuth2Patch struct {
	Enabled                 *bool   `json:"enabled"`
	ClientID                *string `json:"client_id"`
	ClientSecret            *string `json:"client_secret"`
	RedirectURL             *string `json:"redirect_url"`
	AuthURL                 *string `json:"auth_url"`
	TokenURL                *string `json:"token_url"`
	UserInfoURL             *string `json:"user_info_url"`
	AllowNonGitHubEndpoints *bool   `json:"allow_non_github_endpoints"`
}

// Normalize fills defaults for omitted fields after decode.
// Normalize 在解码后为缺省字段填充默认值。
func (s *GithubOAuth2Settings) Normalize() {
	if strings.TrimSpace(s.AuthURL) == "" {
		s.AuthURL = DefaultGitHubAuthURL
	}
	if strings.TrimSpace(s.TokenURL) == "" {
		s.TokenURL = DefaultGitHubTokenURL
	}
	if strings.TrimSpace(s.UserInfoURL) == "" {
		s.UserInfoURL = DefaultGitHubUserInfoURL
	}
}

func validateGithubOAuth2(s *GithubOAuth2Settings) error {
	if !s.Enabled {
		return nil
	}
	if strings.TrimSpace(s.ClientID) == "" {
		return errors.New("github_oauth2.client_id is required when enabled is true")
	}
	if strings.TrimSpace(s.ClientSecret) == "" {
		return errors.New("github_oauth2.client_secret is required when enabled is true")
	}
	if strings.TrimSpace(s.RedirectURL) == "" {
		return errors.New("github_oauth2.redirect_url is required when enabled is true")
	}

	// Validate URL formats for non-empty endpoints.
	// 对非空端点校验 URL 格式。
	for _, endpoint := range []struct {
		name string
		val  string
	}{
		{"redirect_url", s.RedirectURL},
		{"auth_url", s.AuthURL},
		{"token_url", s.TokenURL},
		{"user_info_url", s.UserInfoURL},
	} {
		if endpoint.val == "" {
			continue
		}
		parsed, err := url.Parse(endpoint.val)
		if err != nil {
			return fmt.Errorf("github_oauth2.%s is not a valid URL: %w", endpoint.name, err)
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return fmt.Errorf("github_oauth2.%s must use http or https scheme", endpoint.name)
		}
		if parsed.Host == "" {
			return fmt.Errorf("github_oauth2.%s must be an absolute URL with host", endpoint.name)
		}
	}
	return nil
}
