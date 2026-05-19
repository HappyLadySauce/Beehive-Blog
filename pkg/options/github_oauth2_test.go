package options

import (
	"testing"
)

func validGitHubOAuth2Options() *GithubOAuth2Options {
	return &GithubOAuth2Options{
		Enabled:      true,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost:8080/api/v1/auth/callback",
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		UserInfoURL:  "https://api.github.com/user",
	}
}

func TestGithubOAuth2ValidateSkipsWhenDisabled(t *testing.T) {
	opts := &GithubOAuth2Options{Enabled: false}
	if err := opts.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil when disabled", err)
	}
}

func TestGithubOAuth2ValidateAcceptsDefaultGitHubEndpoints(t *testing.T) {
	if err := validGitHubOAuth2Options().Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestGithubOAuth2ValidateRejectsInsecureGitHubEndpoint(t *testing.T) {
	opts := validGitHubOAuth2Options()
	opts.AuthURL = "http://github.com/login/oauth/authorize"

	if err := opts.Validate(); err == nil {
		t.Fatalf("Validate() error = nil, want error")
	}
}

func TestGithubOAuth2ValidateRejectsNonGitHubEndpointByDefault(t *testing.T) {
	opts := validGitHubOAuth2Options()
	opts.TokenURL = "https://example.com/login/oauth/access_token"

	if err := opts.Validate(); err == nil {
		t.Fatalf("Validate() error = nil, want error")
	}
}

func TestGithubOAuth2ValidateAllowsNonGitHubEndpointWithExplicitOptIn(t *testing.T) {
	opts := validGitHubOAuth2Options()
	opts.TokenURL = "http://127.0.0.1:18080/token"
	opts.AllowNonGitHubEndpoints = true

	if err := opts.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
