package types

import (
	"fmt"
	"strings"
)

// SiteSettings holds public site identity fields persisted in application settings.
// SiteSettings 保存于应用设置中的公开站点身份字段。
type SiteSettings struct {
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

// SitePatch is the partial update payload for public site settings.
// SitePatch 是公开站点设置的部分更新载荷。
type SitePatch struct {
	Name        *string `json:"name"`
	URL         *string `json:"url"`
	Subtitle    *string `json:"subtitle"`
	Description *string `json:"description"`
	Keywords    *string `json:"keywords"`
	LogoURL     *string `json:"logo_url"`
	FaviconURL  *string `json:"favicon_url"`
	ICPBeian    *string `json:"icp_beian"`
	PoliceBeian *string `json:"police_beian"`
	FooterText  *string `json:"footer_text"`
}

func (s *SiteSettings) Normalize() {
	s.Name = strings.TrimSpace(s.Name)
	s.URL = strings.TrimSpace(s.URL)
	s.Subtitle = strings.TrimSpace(s.Subtitle)
	s.Description = strings.TrimSpace(s.Description)
	s.Keywords = strings.TrimSpace(s.Keywords)
	s.LogoURL = strings.TrimSpace(s.LogoURL)
	s.FaviconURL = strings.TrimSpace(s.FaviconURL)
	s.ICPBeian = strings.TrimSpace(s.ICPBeian)
	s.PoliceBeian = strings.TrimSpace(s.PoliceBeian)
	s.FooterText = strings.TrimSpace(s.FooterText)
}

func validateSite(s *SiteSettings) error {
	s.Normalize()
	if len([]rune(s.Name)) > 80 {
		return fmt.Errorf("site name must be at most 80 characters")
	}
	if len([]rune(s.Subtitle)) > 160 {
		return fmt.Errorf("site subtitle must be at most 160 characters")
	}
	if len([]rune(s.Description)) > 1000 {
		return fmt.Errorf("site description must be at most 1000 characters")
	}
	if len([]rune(s.Keywords)) > 512 {
		return fmt.Errorf("site keywords must be at most 512 characters")
	}
	if len([]rune(s.ICPBeian)) > 120 {
		return fmt.Errorf("site icp_beian must be at most 120 characters")
	}
	if len([]rune(s.PoliceBeian)) > 120 {
		return fmt.Errorf("site police_beian must be at most 120 characters")
	}
	if len([]rune(s.FooterText)) > 512 {
		return fmt.Errorf("site footer_text must be at most 512 characters")
	}
	if err := validateOptionalHTTPURL("site url", s.URL); err != nil {
		return err
	}
	if err := validateOptionalHTTPURL("site logo_url", s.LogoURL); err != nil {
		return err
	}
	if err := validateOptionalHTTPURL("site favicon_url", s.FaviconURL); err != nil {
		return err
	}
	return nil
}
