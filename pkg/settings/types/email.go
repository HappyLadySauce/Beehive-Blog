// Package types defines application settings payload shapes.
// Package types 定义应用设置的 payload 结构。
package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
)

// TLS mode constants for outbound SMTP.
// 出站 SMTP 的 TLS 模式常量。
const (
	EmailTLSNone     = "none"
	EmailTLSStartTLS = "starttls"
	EmailTLSDirect   = "tls"
)

// ApplicationSettings is the validated JSON shape stored in setting.application_settings.payload.
// ApplicationSettings 为写入 setting.application_settings.payload 的已校验 JSON 形态。
type ApplicationSettings struct {
	Email EmailSMTPSettings `json:"email"`
}

// EmailSMTPSettings holds SMTP transport options persisted in the database.
// EmailSMTPSettings 保存于数据库的 SMTP 传输选项。
type EmailSMTPSettings struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	FromName string `json:"from_name"`
	TLS      string `json:"tls"`
}

// SettingsPatchRequest is a partial update body; only present top-level keys are merged.
// SettingsPatchRequest 为部分更新请求体；仅出现的顶层键参与合并。
type SettingsPatchRequest struct {
	Email *EmailSMTPPatch `json:"email"`
}

// EmailSMTPPatch uses pointers so omitted JSON fields leave existing values unchanged.
// EmailSMTPPatch 使用指针，JSON 省略的字段保留原值。
type EmailSMTPPatch struct {
	Enabled  *bool   `json:"enabled"`
	Host     *string `json:"host"`
	Port     *int    `json:"port"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	From     *string `json:"from"`
	FromName *string `json:"from_name"`
	TLS      *string `json:"tls"`
}

// DefaultApplicationSettings returns the canonical empty configuration used on first boot.
// DefaultApplicationSettings 返回首次启动用的默认配置。
func DefaultApplicationSettings() ApplicationSettings {
	return ApplicationSettings{
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
	}
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
}

// Validate checks business rules for the full settings document.
// Validate 校验整份设置文档的业务规则。
func (s *ApplicationSettings) Validate() error {
	s.Normalize()
	if err := validateEmailSMTP(&s.Email); err != nil {
		return err
	}
	return nil
}

func validateEmailSMTP(e *EmailSMTPSettings) error {
	tls := strings.TrimSpace(strings.ToLower(e.TLS))
	switch tls {
	case EmailTLSNone, EmailTLSStartTLS, EmailTLSDirect:
	default:
		return fmt.Errorf("email.tls must be one of %q, %q, %q", EmailTLSNone, EmailTLSStartTLS, EmailTLSDirect)
	}
	e.TLS = tls

	if !e.Enabled {
		return nil
	}
	if strings.TrimSpace(e.Host) == "" {
		return errors.New("email.host is required when email.enabled is true")
	}
	if e.Port < 1 || e.Port > 65535 {
		return fmt.Errorf("email.port must be between 1 and 65535, got %d", e.Port)
	}
	from := strings.TrimSpace(e.From)
	if from == "" {
		return errors.New("email.from is required when email.enabled is true")
	}
	if _, err := mail.ParseAddress(from); err != nil {
		return fmt.Errorf("email.from: %w", err)
	}
	return nil
}

// ValidateForSend validates SMTP settings before opening a network connection.
// ValidateForSend 在建立网络连接前校验 SMTP 设置。
func (e EmailSMTPSettings) ValidateForSend() error {
	return validateEmailSMTP(&e)
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
	out.Normalize()
	if err := out.Validate(); err != nil {
		return ApplicationSettings{}, err
	}
	return out, nil
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
