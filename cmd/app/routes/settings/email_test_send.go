package settings

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

const smtpTestTimeout = 10 * time.Second

var sendSMTPTestEmail = sendSMTPTestEmailDefault

// ServeEmailTest sends a test email with the saved SMTP settings.
// ServeEmailTest 使用已保存的 SMTP 设置发送测试邮件。
//
// @Summary      Send SMTP test email (admin)
// @Description  Sends a test email to the requested recipient using saved SMTP settings. 中文：使用已保存 SMTP 设置向指定收件人发送测试邮件。
// @Tags         settings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      v1.SettingsEmailTestRequest  true  "Test recipient"
// @Success      200   {object}  common.BaseResponse{data=v1.SettingsEmailTestResponse}
// @Failure      400   {object}  common.BaseResponse
// @Failure      401   {object}  common.BaseResponse
// @Failure      403   {object}  common.BaseResponse
// @Failure      500   {object}  common.BaseResponse
// @Router       /api/v1/settings/email/test [post]
func (h *SettingsController) TestEmail(ctx *gin.Context) {
	if h.svc.Settings == nil {
		common.Fail(ctx, common.NewInternal("settings provider is not configured", fmt.Errorf("nil settings provider")))
		return
	}
	var req v1.SettingsEmailTestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	recipient := strings.TrimSpace(req.Recipient)
	settings := h.svc.Settings.Current()
	if err := validateEmailTestSettings(settings.Email, recipient); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid email test settings", err))
		return
	}
	if err := sendSMTPTestEmail(ctx.Request.Context(), settings.Email, recipient); err != nil {
		common.Fail(ctx, common.NewInternal("failed to send test email", err))
		return
	}
	common.Success(ctx, v1.SettingsEmailTestResponse{Recipient: recipient})
}

func validateEmailTestSettings(email settingtypes.EmailSMTPSettings, recipient string) error {
	if !email.Enabled {
		return fmt.Errorf("email sending is disabled")
	}
	if _, err := mail.ParseAddress(recipient); err != nil {
		return fmt.Errorf("recipient: %w", err)
	}
	if strings.TrimSpace(email.Username) != "" && strings.TrimSpace(email.Password) == "" {
		return fmt.Errorf("email.password is required when email.username is set")
	}
	return email.ValidateForSend()
}

func sendSMTPTestEmailDefault(ctx context.Context, cfg settingtypes.EmailSMTPSettings, recipient string) error {
	ctx, cancel := context.WithTimeout(ctx, smtpTestTimeout)
	defer cancel()

	host := strings.TrimSpace(cfg.Host)
	tlsMode := strings.TrimSpace(strings.ToLower(cfg.TLS))
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", cfg.Port))
	from := strings.TrimSpace(cfg.From)
	fromName := strings.TrimSpace(cfg.FromName)
	fromHeader := from
	if fromName != "" {
		fromHeader = (&mail.Address{Name: fromName, Address: from}).String()
	}
	message := []byte(strings.Join([]string{
		"From: " + fromHeader,
		"To: " + recipient,
		"Subject: Beehive Blog SMTP test",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		"Beehive Blog SMTP test email was sent successfully.",
	}, "\r\n"))

	var client *smtp.Client
	var err error
	dialer := &net.Dialer{Timeout: smtpTestTimeout}
	deadline := time.Now().Add(smtpTestTimeout)
	if tlsMode == settingtypes.EmailTLSDirect {
		conn, dialErr := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
		if dialErr != nil {
			return dialErr
		}
		_ = conn.SetDeadline(deadline)
		client, err = smtp.NewClient(conn, host)
	} else {
		conn, dialErr := dialer.DialContext(ctx, "tcp", addr)
		if dialErr != nil {
			return dialErr
		}
		_ = conn.SetDeadline(deadline)
		client, err = smtp.NewClient(conn, host)
	}
	if err != nil {
		return err
	}
	defer client.Close()

	if tlsMode == settingtypes.EmailTLSStartTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("smtp server does not advertise STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if strings.TrimSpace(cfg.Username) != "" {
		auth := smtp.PlainAuth("", strings.TrimSpace(cfg.Username), cfg.Password, host)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(recipient); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}
