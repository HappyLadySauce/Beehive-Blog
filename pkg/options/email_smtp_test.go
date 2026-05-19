package options

import (
	"testing"

	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

func TestEmailSMTPOptionsValidateDisabledAllowsEmptyHost(t *testing.T) {
	e := NewEmailSMTPOptions()
	if err := e.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestEmailSMTPOptionsValidateEnabledCompleteConfigRequiresValidFrom(t *testing.T) {
	e := NewEmailSMTPOptions()
	e.Enabled = true
	e.Host = "smtp.example.com"
	e.From = "not-an-email"
	if err := e.Validate(); err == nil {
		t.Fatal("expected error for invalid from when enabled config is complete")
	}
}

func TestEmailSMTPOptionsValidateEnabledPlaceholderAllowsMissingHostAndFrom(t *testing.T) {
	e := NewEmailSMTPOptions()
	e.Enabled = true

	got, err := e.ToApplicationSettings()
	if err != nil {
		t.Fatalf("ToApplicationSettings: %v", err)
	}
	if got.Email.Enabled {
		t.Fatalf("Email.Enabled: got true, want false for incomplete startup placeholder")
	}
	if err := e.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestEmailSMTPOptionsValidateEnabledPlaceholderAllowsPartiallyMissingHostOrFrom(t *testing.T) {
	tests := []struct {
		name string
		host string
		from string
	}{
		{name: "missing host", host: "", from: "robot@example.com"},
		{name: "missing from", host: "smtp.example.com", from: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEmailSMTPOptions()
			e.Enabled = true
			e.Host = tt.host
			e.From = tt.from

			got, err := e.ToApplicationSettings()
			if err != nil {
				t.Fatalf("ToApplicationSettings: %v", err)
			}
			if got.Email.Enabled {
				t.Fatalf("Email.Enabled: got true, want false for incomplete startup placeholder")
			}
			if got.Email.Host != tt.host || got.Email.From != tt.from {
				t.Fatalf("Email fields not preserved: %+v", got.Email)
			}
		})
	}
}

func TestEmailSMTPOptionsValidateEnabledWithHostAndFromStaysEnabled(t *testing.T) {
	e := NewEmailSMTPOptions()
	e.Enabled = true
	e.Host = "smtp.example.com"
	e.From = "robot@example.com"

	got, err := e.ToApplicationSettings()
	if err != nil {
		t.Fatalf("ToApplicationSettings: %v", err)
	}
	if !got.Email.Enabled {
		t.Fatal("Email.Enabled: got false, want true for complete SMTP config")
	}
}

func TestEmailSMTPOptionsValidateInvalidTLS(t *testing.T) {
	e := NewEmailSMTPOptions()
	e.TLS = "nope"
	if err := e.Validate(); err == nil {
		t.Fatal("expected error for invalid tls")
	}
}

func TestEmailSMTPOptionsToApplicationSettings(t *testing.T) {
	e := NewEmailSMTPOptions()
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
