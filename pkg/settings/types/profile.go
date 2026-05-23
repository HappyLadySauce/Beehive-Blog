package types

import (
	"fmt"
	"net/url"
	"strings"
)

// ProfileSettings holds public author profile fields persisted in application settings.
// ProfileSettings 保存于应用设置中的公开博主资料字段。
type ProfileSettings struct {
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Headline    string `json:"headline"`
	Bio         string `json:"bio"`
	Location    string `json:"location"`
	Website     string `json:"website"`
}

// ProfilePatch is the partial update payload for public author profile settings.
// ProfilePatch 是公开博主资料设置的部分更新载荷。
type ProfilePatch struct {
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
	Headline    *string `json:"headline"`
	Bio         *string `json:"bio"`
	Location    *string `json:"location"`
	Website     *string `json:"website"`
}

func (p *ProfileSettings) Normalize() {
	p.DisplayName = strings.TrimSpace(p.DisplayName)
	p.AvatarURL = strings.TrimSpace(p.AvatarURL)
	p.Headline = strings.TrimSpace(p.Headline)
	p.Bio = strings.TrimSpace(p.Bio)
	p.Location = strings.TrimSpace(p.Location)
	p.Website = strings.TrimSpace(p.Website)
}

func validateProfile(p *ProfileSettings) error {
	p.Normalize()
	if len([]rune(p.DisplayName)) > 80 {
		return fmt.Errorf("profile display_name must be at most 80 characters")
	}
	if len([]rune(p.Headline)) > 160 {
		return fmt.Errorf("profile headline must be at most 160 characters")
	}
	if len([]rune(p.Bio)) > 1000 {
		return fmt.Errorf("profile bio must be at most 1000 characters")
	}
	if len([]rune(p.Location)) > 120 {
		return fmt.Errorf("profile location must be at most 120 characters")
	}
	if err := validateOptionalHTTPURL("profile avatar_url", p.AvatarURL); err != nil {
		return err
	}
	if err := validateOptionalHTTPURL("profile website", p.Website); err != nil {
		return err
	}
	return nil
}

func validateOptionalHTTPURL(label string, raw string) error {
	if raw == "" {
		return nil
	}
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("%s must be a valid URL", label)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("%s must use http or https", label)
	}
	return nil
}
