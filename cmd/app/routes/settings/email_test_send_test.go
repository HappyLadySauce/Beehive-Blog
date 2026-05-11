package settings_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

func TestEmailTestRejectsDisabledSMTP(t *testing.T) {
	email := validEmailTestSettings()
	email.Enabled = false

	rec, envelope := performTestEmail(t, email, "reader@example.com")
	assertEmailTestError(t, rec, envelope, http.StatusBadRequest, "invalid email test settings")
}

func TestEmailTestRejectsInvalidRecipient(t *testing.T) {
	email := validEmailTestSettings()

	rec, envelope := performTestEmail(t, email, "not-an-email")
	assertEmailTestError(t, rec, envelope, http.StatusBadRequest, "invalid request body")
}

func TestEmailTestRejectsMissingPasswordWhenUsernameIsSet(t *testing.T) {
	email := validEmailTestSettings()
	email.Password = ""

	rec, envelope := performTestEmail(t, email, "reader@example.com")
	assertEmailTestError(t, rec, envelope, http.StatusBadRequest, "invalid email test settings")
}

func TestEmailTestAllowsUnauthenticatedSMTP(t *testing.T) {
	host, port := startFakeSMTPServer(t)
	email := validEmailTestSettings()
	email.Host = host
	email.Port = port
	email.Username = ""
	email.Password = ""
	email.TLS = settingtypes.EmailTLSNone

	rec, envelope := performTestEmail(t, email, "reader@example.com")
	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if envelope.Code != http.StatusOK {
		t.Fatalf("envelope code = %d, want %d", envelope.Code, http.StatusOK)
	}
	if envelope.Data.Recipient != "reader@example.com" {
		t.Fatalf("recipient = %q, want %q", envelope.Data.Recipient, "reader@example.com")
	}
}

type emailTestEnvelope struct {
	Code int                          `json:"code"`
	Msg  string                       `json:"message"`
	Data v1.SettingsEmailTestResponse `json:"data"`
}

func validEmailTestSettings() settingtypes.EmailSMTPSettings {
	return settingtypes.EmailSMTPSettings{
		Enabled:  true,
		Host:     "smtp.example.com",
		Port:     587,
		Username: "robot",
		Password: "secret",
		From:     "robot@example.com",
		FromName: "Beehive",
		TLS:      settingtypes.EmailTLSStartTLS,
	}
}

func performTestEmail(t *testing.T, email settingtypes.EmailSMTPSettings, recipient string) (*httptest.ResponseRecorder, emailTestEnvelope) {
	t.Helper()
	s := settingtypes.DefaultApplicationSettings()
	s.Email = email
	c, mock := newSettingsTestController(t, s, 1)

	body, err := json.Marshal(v1.SettingsEmailTestRequest{Recipient: recipient})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/settings/email/test", bytes.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	c.TestEmail(ctx)

	var envelope emailTestEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
	return rec, envelope
}

func assertEmailTestError(t *testing.T, rec *httptest.ResponseRecorder, envelope emailTestEnvelope, status int, message string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("HTTP status = %d, want %d", rec.Code, status)
	}
	if envelope.Code != status {
		t.Fatalf("envelope code = %d, want %d", envelope.Code, status)
	}
	if envelope.Msg != message {
		t.Fatalf("message = %q, want %q", envelope.Msg, message)
	}
}

func startFakeSMTPServer(t *testing.T) (string, int) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen fake smtp: %v", err)
	}
	t.Cleanup(func() {
		_ = ln.Close()
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

		reader := bufio.NewReader(conn)
		writeSMTPLine(t, conn, "220 beehive-test ESMTP")
		inData := false
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if inData {
				if line == "." {
					inData = false
					writeSMTPLine(t, conn, "250 ok")
				}
				continue
			}
			upper := strings.ToUpper(line)
			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				writeSMTPLine(t, conn, "250 beehive-test")
			case strings.HasPrefix(upper, "MAIL FROM:"), strings.HasPrefix(upper, "RCPT TO:"):
				writeSMTPLine(t, conn, "250 ok")
			case upper == "DATA":
				inData = true
				writeSMTPLine(t, conn, "354 end data with <CR><LF>.<CR><LF>")
			case upper == "QUIT":
				writeSMTPLine(t, conn, "221 bye")
				return
			default:
				writeSMTPLine(t, conn, "250 ok")
			}
		}
	}()
	t.Cleanup(func() {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatalf("fake smtp server did not stop")
		}
	})

	host, portText, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("split listener address: %v", err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatalf("parse listener port: %v", err)
	}
	return host, port
}

func writeSMTPLine(t *testing.T, conn net.Conn, line string) {
	t.Helper()
	if _, err := conn.Write([]byte(line + "\r\n")); err != nil {
		t.Fatalf("write fake smtp response: %v", err)
	}
}
