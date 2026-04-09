// Package mailer 提供基于 net/smtp 标准库的 SMTP 邮件发送能力。
// 设计原则：
//   - 零额外依赖（仅 net/smtp + crypto/tls）
//   - 配置懒加载：从 settings 表 group=smtp 读取，未配置时 Mailer 为 nil，调用方跳过发送
//   - Send 方法幂等、带 context 超时控制
package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Config SMTP 配置（对应 settings 表 group=smtp 的各 key）。
type Config struct {
	Host     string // smtp.host
	Port     string // smtp.port，默认 "465"
	Username string // smtp.username（发件人邮箱）
	Password string // smtp.password
	FromName string // smtp.fromName，默认同 Username
	// Encryption: ssl（隐式 TLS / SMTPS，常见 465）| tls（明文后 STARTTLS，常见 587）| none（全程明文）
	Encryption string // smtp.encryption，默认 "ssl"
}

// IsValid 检查最低必填字段。
func (c *Config) IsValid() bool {
	return strings.TrimSpace(c.Host) != "" &&
		strings.TrimSpace(c.Username) != "" &&
		strings.TrimSpace(c.Password) != ""
}

// SMTPMailer 使用 net/smtp 发送 HTML 邮件。
type SMTPMailer struct {
	cfg Config
}

// New 构建 SMTPMailer；cfg.IsValid() 为 false 时返回 nil, error。
func New(cfg Config) (*SMTPMailer, error) {
	if !cfg.IsValid() {
		return nil, errors.New("mailer: incomplete SMTP configuration (host/username/password required)")
	}
	if cfg.Port == "" {
		cfg.Port = "465"
	}
	if cfg.Encryption == "" {
		cfg.Encryption = "ssl"
	}
	if cfg.FromName == "" {
		cfg.FromName = cfg.Username
	}
	return &SMTPMailer{cfg: cfg}, nil
}

// Send 发送一封 HTML 邮件。to 为收件人邮箱，subject 为主题，htmlBody 为 HTML 正文。
// 遵守 ctx 超时：连接建立时使用 ctx 派生的 deadline。
func (m *SMTPMailer) Send(ctx context.Context, to, subject, htmlBody string) error {
	if m == nil {
		return errors.New("mailer: not configured")
	}
	to = strings.TrimSpace(to)
	if to == "" {
		return errors.New("mailer: recipient address is empty")
	}

	addr := net.JoinHostPort(m.cfg.Host, m.cfg.Port)
	from := fmt.Sprintf("%s <%s>", m.cfg.FromName, m.cfg.Username)

	msg := buildMIMEMessage(m.cfg.Username, to, from, subject, htmlBody)

	// 使用 ctx 中的 deadline 控制连接超时
	dialer := &net.Dialer{}
	if dl, ok := ctx.Deadline(); ok {
		dialer.Deadline = dl
	} else {
		dialer.Timeout = 15 * time.Second
	}

	enc := strings.ToLower(strings.TrimSpace(m.cfg.Encryption))
	switch enc {
	case "ssl":
		return m.sendImplicitTLS(dialer, addr, to, msg)
	case "tls":
		return m.sendSTARTTLS(dialer, addr, to, msg)
	default:
		return m.sendPlain(dialer, addr, to, msg)
	}
}

// sendImplicitTLS 连接即 TLS（SMTPS），适用于常见 465 端口。
func (m *SMTPMailer) sendImplicitTLS(dialer *net.Dialer, addr, to string, msg []byte) error {
	tlsCfg := &tls.Config{
		ServerName: m.cfg.Host,
		MinVersion: tls.VersionTLS12,
	}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("mailer: TLS dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		return fmt.Errorf("mailer: SMTP client init failed: %w", err)
	}
	defer client.Quit() //nolint:errcheck

	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("mailer: SMTP auth failed: %w", err)
	}
	return sendViaClient(client, m.cfg.Username, to, msg)
}

// sendSTARTTLS 先明文连接再 STARTTLS 升级，适用于常见 587 端口。
func (m *SMTPMailer) sendSTARTTLS(dialer *net.Dialer, addr, to string, msg []byte) error {
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("mailer: dial failed: %w", err)
	}

	client, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("mailer: SMTP client init failed: %w", err)
	}
	defer client.Quit() //nolint:errcheck

	tlsCfg := &tls.Config{
		ServerName: m.cfg.Host,
		MinVersion: tls.VersionTLS12,
	}
	if err := client.StartTLS(tlsCfg); err != nil {
		return fmt.Errorf("mailer: STARTTLS failed: %w", err)
	}

	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("mailer: SMTP auth failed: %w", err)
	}
	return sendViaClient(client, m.cfg.Username, to, msg)
}

func (m *SMTPMailer) sendPlain(dialer *net.Dialer, addr, to string, msg []byte) error {
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("mailer: dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		return fmt.Errorf("mailer: SMTP client init failed: %w", err)
	}
	defer client.Quit() //nolint:errcheck

	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("mailer: SMTP auth failed: %w", err)
	}
	return sendViaClient(client, m.cfg.Username, to, msg)
}

func sendViaClient(client *smtp.Client, from, to string, msg []byte) error {
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("mailer: MAIL FROM failed: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("mailer: RCPT TO failed: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("mailer: DATA failed: %w", err)
	}
	defer w.Close()
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("mailer: write message failed: %w", err)
	}
	return nil
}

// buildMIMEMessage 构建符合 RFC 2822 的 MIME 邮件字节。
func buildMIMEMessage(fromAddr, toAddr, fromDisplay, subject, htmlBody string) []byte {
	var sb strings.Builder
	sb.WriteString("From: " + fromDisplay + "\r\n")
	sb.WriteString("To: " + toAddr + "\r\n")
	sb.WriteString("Subject: " + subject + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	sb.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(htmlBody)
	_ = fromAddr // used via fromDisplay
	return []byte(sb.String())
}
