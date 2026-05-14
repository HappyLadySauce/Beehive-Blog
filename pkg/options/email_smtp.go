package options

import (
	"strings"

	"github.com/spf13/pflag"

	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

// EmailSMTPOptions holds CLI / file / env knobs for outbound SMTP defaults and DB seeding.
// EmailSMTPOptions 保存出站 SMTP 的 CLI、文件与环境变量默认值，并用于数据库补种。
type EmailSMTPOptions struct {
	// Enabled turns on outbound SMTP when credentials are valid.
	// Enabled 在凭据有效时启用出站 SMTP。
	Enabled bool `json:"enabled" mapstructure:"enabled"`
	// Host is the SMTP server hostname.
	// Host 为 SMTP 服务器主机名。
	Host string `json:"host" mapstructure:"host"`
	// Port is the SMTP server TCP port.
	// Port 为 SMTP 服务器 TCP 端口。
	Port int `json:"port" mapstructure:"port"`
	// Username is the SMTP AUTH user, if any.
	// Username 为 SMTP AUTH 用户名（可空）。
	Username string `json:"username" mapstructure:"username"`
	// Password is the SMTP AUTH secret; excluded from JSON dumps.
	// Password 为 SMTP AUTH 密码；不在 JSON 调试输出中出现。
	Password string `json:"-" mapstructure:"password"`
	// From is the RFC5322 From address when SMTP is enabled.
	// From 为启用 SMTP 时的 RFC5322 From 地址。
	From string `json:"from" mapstructure:"from"`
	// FromName is the optional display name for the From header.
	// FromName 为 From 头的可选显示名。
	FromName string `json:"from-name" mapstructure:"from-name"`
	// TLS selects none|starttls|tls for transport security.
	// TLS 选择 none|starttls|tls 传输安全模式。
	TLS string `json:"tls" mapstructure:"tls"`
}

// NewEmailSMTPOptions returns defaults aligned with settingtypes.DefaultApplicationSettings.
// NewEmailSMTPOptions 返回与 settingtypes.DefaultApplicationSettings 一致的默认值。
func NewEmailSMTPOptions() *EmailSMTPOptions {
	def := settingtypes.DefaultApplicationSettings().Email
	return &EmailSMTPOptions{
		Enabled:  def.Enabled,
		Host:     def.Host,
		Port:     def.Port,
		Username: def.Username,
		Password: def.Password,
		From:     def.From,
		FromName: def.FromName,
		TLS:      def.TLS,
	}
}

// ToApplicationSettings maps file/CLI/env options into the persisted settings shape and validates it.
// When email is enabled but host and from are both empty, outbound mail cannot run; treat as seed placeholder and persist enabled=false.
// ToApplicationSettings 将文件/CLI/环境选项映射为持久化设置形态并校验。
// 若启用邮件但 host 与 from 皆空，出站无法投递；视为补种占位并持久化为 enabled=false。
func (e *EmailSMTPOptions) ToApplicationSettings() (settingtypes.ApplicationSettings, error) {
	if e == nil {
		s := settingtypes.DefaultApplicationSettings()
		return s, s.Validate()
	}
	s := settingtypes.ApplicationSettings{
		Email: settingtypes.EmailSMTPSettings{
			Enabled:  e.Enabled,
			Host:     e.Host,
			Port:     e.Port,
			Username: e.Username,
			Password: e.Password,
			From:     e.From,
			FromName: e.FromName,
			TLS:      e.TLS,
		},
	}
	s.Normalize()
	if s.Email.Enabled && (strings.TrimSpace(s.Email.Host) == "" || strings.TrimSpace(s.Email.From) == "") {
		s.Email.Enabled = false
	}
	if err := s.Validate(); err != nil {
		return settingtypes.ApplicationSettings{}, err
	}
	return s, nil
}

// Validate checks SMTP options using the same rules as stored application settings.
// Validate 使用与已存储应用设置相同的规则校验 SMTP 选项。
func (e *EmailSMTPOptions) Validate() error {
	_, err := e.ToApplicationSettings()
	return err
}

// AddFlags registers email SMTP flags on the supplied FlagSet.
// AddFlags 将邮箱 SMTP 相关命令行标志注册到给定的 FlagSet。
func (e *EmailSMTPOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&e.Enabled, "email-enabled", e.Enabled, "SMTP: enable sending when host/from are set; optional placeholder when both empty (seed)")
	fs.StringVar(&e.Host, "email-host", e.Host, "SMTP server hostname")
	fs.IntVar(&e.Port, "email-port", e.Port, "SMTP server TCP port")
	fs.StringVar(&e.Username, "email-username", e.Username, "SMTP AUTH username (optional)")
	fs.StringVar(&e.Password, "email-password", e.Password, "SMTP AUTH password (optional; excluded from options JSON dump)")
	fs.StringVar(&e.From, "email-from", e.From, "RFC5322 From address (required when email-enabled and host or from is non-empty)")
	fs.StringVar(&e.FromName, "email-from-name", e.FromName, "Optional display name for From header")
	fs.StringVar(&e.TLS, "email-tls", e.TLS, "TLS mode: none, starttls, or tls")
}
