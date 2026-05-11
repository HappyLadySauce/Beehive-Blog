package options_test

import (
	"testing"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

func TestEmailSMTPOptionsValidateDisabledAllowsEmptyHost(t *testing.T) {
	e := options.NewEmailSMTPOptions()
	if err := e.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestEmailSMTPOptionsValidateEnabledRequiresHostAndFrom(t *testing.T) {
	e := options.NewEmailSMTPOptions()
	e.Enabled = true
	e.Host = ""
	e.From = "a@b.com"
	if err := e.Validate(); err == nil {
		t.Fatal("expected error for empty host when enabled")
	}

	e = options.NewEmailSMTPOptions()
	e.Enabled = true
	e.Host = "smtp.example.com"
	e.From = ""
	if err := e.Validate(); err == nil {
		t.Fatal("expected error for empty from when enabled")
	}
}

func TestEmailSMTPOptionsValidateInvalidTLS(t *testing.T) {
	e := options.NewEmailSMTPOptions()
	e.TLS = "nope"
	if err := e.Validate(); err == nil {
		t.Fatal("expected error for invalid tls")
	}
}

func TestEmailSMTPOptionsToApplicationSettings(t *testing.T) {
	e := options.NewEmailSMTPOptions()
	e.Host = "smtp.example.com"
	e.From = "noreply@example.com"
	got, err := e.ToApplicationSettings()
	if err != nil {
		t.Fatalf("ToApplicationSettings: %v", err)
	}
	if got.Email.Host != e.Host || got.Email.From != e.From {
		t.Fatalf("unexpected mapping: %+v", got.Email)
	}
	if got.Email.TLS != settingtypes.EmailTLSStartTLS {
		t.Fatalf("tls = %q", got.Email.TLS)
	}
}
