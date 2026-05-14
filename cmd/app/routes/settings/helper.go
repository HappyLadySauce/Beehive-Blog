package settings

import (
	"strings"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

// toResponse maps internal settings to a sanitized API response.
// toResponse 将内部设置映射为脱敏 API 响应。
func toResponse(s settingtypes.ApplicationSettings, revision int64) v1.SettingsResponse {
	e := s.Email
	pwdSet := strings.TrimSpace(e.Password) != ""

	g := s.GithubOAuth2
	secretSet := strings.TrimSpace(g.ClientSecret) != ""

	return v1.SettingsResponse{
		Revision: revision,
		Email: v1.EmailSettingsPublic{
			Enabled:     e.Enabled,
			Host:        e.Host,
			Port:        e.Port,
			Username:    e.Username,
			PasswordSet: pwdSet,
			From:        e.From,
			FromName:    e.FromName,
			TLS:         e.TLS,
		},
		GithubOAuth2: v1.GithubOAuth2SettingsPublic{
			Enabled:                 g.Enabled,
			ClientID:                g.ClientID,
			ClientSecretSet:         secretSet,
			RedirectURL:             g.RedirectURL,
			AuthURL:                 g.AuthURL,
			TokenURL:                g.TokenURL,
			UserInfoURL:             g.UserInfoURL,
			AllowNonGitHubEndpoints: g.AllowNonGitHubEndpoints,
		},
	}
}

func patchFromV1(p *v1.EmailSMTPPatchJSON) *settingtypes.EmailSMTPPatch {
	if p == nil {
		return nil
	}
	return &settingtypes.EmailSMTPPatch{
		Enabled:  p.Enabled,
		Host:     p.Host,
		Port:     p.Port,
		Username: p.Username,
		Password: p.Password,
		From:     p.From,
		FromName: p.FromName,
		TLS:      p.TLS,
	}
}

func patchGithubFromV1(p *v1.GithubOAuth2PatchJSON) *settingtypes.GithubOAuth2Patch {
	if p == nil {
		return nil
	}
	return &settingtypes.GithubOAuth2Patch{
		Enabled:                 p.Enabled,
		ClientID:                p.ClientID,
		ClientSecret:            p.ClientSecret,
		RedirectURL:             p.RedirectURL,
		AuthURL:                 p.AuthURL,
		TokenURL:                p.TokenURL,
		UserInfoURL:             p.UserInfoURL,
		AllowNonGitHubEndpoints: p.AllowNonGitHubEndpoints,
	}
}
