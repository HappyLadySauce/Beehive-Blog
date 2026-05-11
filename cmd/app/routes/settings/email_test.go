package settings

import (
	"testing"

	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

func TestToResponsePasswordSet(t *testing.T) {
	s := settingtypes.DefaultApplicationSettings()
	s.Email.Password = "secret"
	r := toResponse(s, 7)
	if !r.Email.PasswordSet {
		t.Fatal("expected PasswordSet true")
	}
	if r.Revision != 7 {
		t.Fatalf("revision = %d", r.Revision)
	}
}

func TestValidateEmailTestSettingsRequiresEnabledSMTP(t *testing.T) {
	email := validEmailTestSettings()
	email.Enabled = false

	if err := validateEmailTestSettings(email, "reader@example.com"); err == nil {
		t.Fatal("expected disabled SMTP to be rejected")
	}
}

func TestValidateEmailTestSettingsRequiresValidRecipient(t *testing.T) {
	email := validEmailTestSettings()

	if err := validateEmailTestSettings(email, "not-an-email"); err == nil {
		t.Fatal("expected invalid recipient to be rejected")
	}
}

func TestValidateEmailTestSettingsRequiresPasswordWhenUsernameIsSet(t *testing.T) {
	email := validEmailTestSettings()
	email.Password = ""

	if err := validateEmailTestSettings(email, "reader@example.com"); err == nil {
		t.Fatal("expected missing password to be rejected when username is set")
	}
}

func TestValidateEmailTestSettingsAllowsUnauthenticatedSMTP(t *testing.T) {
	email := validEmailTestSettings()
	email.Username = ""
	email.Password = ""

	if err := validateEmailTestSettings(email, "reader@example.com"); err != nil {
		t.Fatalf("expected unauthenticated SMTP to pass validation, got %v", err)
	}
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
