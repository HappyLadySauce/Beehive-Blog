// Package types defines application settings payload shapes.
// Package types 定义应用设置的 payload 结构。
package types

import (
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
