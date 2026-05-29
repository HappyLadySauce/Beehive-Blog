package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ApplicationSettings is the validated JSON shape stored in setting.application_settings.payload.
// ApplicationSettings 为写入 setting.application_settings.payload 的已校验 JSON 形态。
type ApplicationSettings struct {
	Email        EmailSMTPSettings    `json:"email"`
	GithubOAuth2 GithubOAuth2Settings `json:"github_oauth2"`
	Profile      ProfileSettings      `json:"profile"`
	Site         SiteSettings         `json:"site"`
}

// Normalize fills defaults for omitted JSON fields after decode.
// Normalize 在解码后为缺省字段填充默认值。
func (s *ApplicationSettings) Normalize() {
	if s.Email.Port <= 0 {
		s.Email.Port = 587
	}
	if strings.TrimSpace(s.Email.TLS) == "" {
		s.Email.TLS = EmailTLSStartTLS
	}
	s.GithubOAuth2.Normalize()
	s.Profile.Normalize()
	s.Site.Normalize()
}

// Validate checks business rules for the full settings document.
// Validate 校验整份设置文档的业务规则。
func (s *ApplicationSettings) Validate() error {
	s.Normalize()
	if err := validateEmailSMTP(&s.Email); err != nil {
		return err
	}
	if err := validateGithubOAuth2(&s.GithubOAuth2); err != nil {
		return err
	}
	if err := validateProfile(&s.Profile); err != nil {
		return err
	}
	if err := validateSite(&s.Site); err != nil {
		return err
	}
	return nil
}

// DefaultApplicationSettings returns the canonical empty configuration used on first boot.
// DefaultApplicationSettings 返回首次启动用的默认配置。
func DefaultApplicationSettings() ApplicationSettings {
	s := ApplicationSettings{
		Email: EmailSMTPSettings{
			Enabled:  false,
			Host:     "",
			Port:     587,
			Username: "",
			Password: "",
			From:     "",
			FromName: "",
			TLS:      EmailTLSStartTLS,
		},
		GithubOAuth2: GithubOAuth2Settings{
			Enabled: false,
		},
		Profile: ProfileSettings{
			DisplayName: "Beehive",
			Headline:    "个人博客与知识中台",
		},
		Site: SiteSettings{
			Name:        "Beehive",
			Subtitle:    "Beehive Blog",
			Description: "个人博客、AI 协作创作与面向智能体的个人知识中台。",
		},
	}
	s.GithubOAuth2.Normalize()
	return s
}

// MergePatch merges a patch into a deep copy of base and returns the result.
// MergePatch 将补丁合并到 base 的深拷贝并返回结果。
func MergePatch(base ApplicationSettings, patch *SettingsPatchRequest) (ApplicationSettings, error) {
	if patch == nil {
		return base, errors.New("patch is nil")
	}
	out := base
	if patch.Email != nil {
		p := patch.Email
		if p.Enabled != nil {
			out.Email.Enabled = *p.Enabled
		}
		if p.Host != nil {
			out.Email.Host = *p.Host
		}
		if p.Port != nil {
			out.Email.Port = *p.Port
		}
		if p.Username != nil {
			out.Email.Username = *p.Username
		}
		if p.Password != nil {
			out.Email.Password = *p.Password
		}
		if p.From != nil {
			out.Email.From = *p.From
		}
		if p.FromName != nil {
			out.Email.FromName = *p.FromName
		}
		if p.TLS != nil {
			out.Email.TLS = *p.TLS
		}
	}
	if patch.GithubOAuth2 != nil {
		p := patch.GithubOAuth2
		if p.Enabled != nil {
			out.GithubOAuth2.Enabled = *p.Enabled
		}
		if p.ClientID != nil {
			out.GithubOAuth2.ClientID = *p.ClientID
		}
		if p.ClientSecret != nil {
			out.GithubOAuth2.ClientSecret = *p.ClientSecret
		}
		if p.RedirectURL != nil {
			out.GithubOAuth2.RedirectURL = *p.RedirectURL
		}
		if p.AuthURL != nil {
			out.GithubOAuth2.AuthURL = *p.AuthURL
		}
		if p.TokenURL != nil {
			out.GithubOAuth2.TokenURL = *p.TokenURL
		}
		if p.UserInfoURL != nil {
			out.GithubOAuth2.UserInfoURL = *p.UserInfoURL
		}
		if p.AllowNonGitHubEndpoints != nil {
			out.GithubOAuth2.AllowNonGitHubEndpoints = *p.AllowNonGitHubEndpoints
		}
	}
	if patch.Profile != nil {
		p := patch.Profile
		if p.DisplayName != nil {
			out.Profile.DisplayName = *p.DisplayName
		}
		if p.AvatarURL != nil {
			out.Profile.AvatarURL = *p.AvatarURL
		}
		if p.Headline != nil {
			out.Profile.Headline = *p.Headline
		}
		if p.Bio != nil {
			out.Profile.Bio = *p.Bio
		}
		if p.Location != nil {
			out.Profile.Location = *p.Location
		}
		if p.Website != nil {
			out.Profile.Website = *p.Website
		}
	}
	if patch.Site != nil {
		p := patch.Site
		if p.Name != nil {
			out.Site.Name = *p.Name
		}
		if p.URL != nil {
			out.Site.URL = *p.URL
		}
		if p.Subtitle != nil {
			out.Site.Subtitle = *p.Subtitle
		}
		if p.Description != nil {
			out.Site.Description = *p.Description
		}
		if p.Keywords != nil {
			out.Site.Keywords = *p.Keywords
		}
		if p.LogoURL != nil {
			out.Site.LogoURL = *p.LogoURL
		}
		if p.FaviconURL != nil {
			out.Site.FaviconURL = *p.FaviconURL
		}
		if p.ICPBeian != nil {
			out.Site.ICPBeian = *p.ICPBeian
		}
		if p.PoliceBeian != nil {
			out.Site.PoliceBeian = *p.PoliceBeian
		}
		if p.FooterText != nil {
			out.Site.FooterText = *p.FooterText
		}
	}
	out.Normalize()
	if err := out.Validate(); err != nil {
		return ApplicationSettings{}, err
	}
	return out, nil
}

// SettingsPatchRequest is a partial update body; only present top-level keys are merged.
// SettingsPatchRequest 为部分更新请求体；仅出现的顶层键参与合并。
type SettingsPatchRequest struct {
	Email        *EmailSMTPPatch    `json:"email"`
	GithubOAuth2 *GithubOAuth2Patch `json:"github_oauth2"`
	Profile      *ProfilePatch      `json:"profile"`
	Site         *SitePatch         `json:"site"`
}

// ParsePayload decodes JSON bytes into ApplicationSettings and validates.
// ParsePayload 将 JSON 字节解码为 ApplicationSettings 并校验。
func ParsePayload(raw []byte) (ApplicationSettings, error) {
	if len(raw) == 0 || string(raw) == "null" {
		s := DefaultApplicationSettings()
		return s, s.Validate()
	}
	var s ApplicationSettings
	if err := json.Unmarshal(raw, &s); err != nil {
		return ApplicationSettings{}, fmt.Errorf("decode settings payload: %w", err)
	}
	s.Normalize()
	if err := s.Validate(); err != nil {
		return ApplicationSettings{}, err
	}
	return s, nil
}

// MarshalPayload serializes settings to JSON for persistence.
// MarshalPayload 将设置序列化为 JSON 以便持久化。
func MarshalPayload(s ApplicationSettings) ([]byte, error) {
	return json.Marshal(s)
}
