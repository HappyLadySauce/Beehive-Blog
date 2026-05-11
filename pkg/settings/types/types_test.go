package types_test

import (
	"strings"
	"testing"

	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

func TestValidateEmailDisabledSkipsHost(t *testing.T) {
	s := settingtypes.DefaultApplicationSettings()
	if err := s.Validate(); err != nil {
		t.Fatalf("Validate() = %v", err)
	}
}

func TestValidateEmailEnabledRequiresHost(t *testing.T) {
	s := settingtypes.DefaultApplicationSettings()
	s.Email.Enabled = true
	s.Email.Host = ""
	if err := s.Validate(); err == nil || !strings.Contains(err.Error(), "host") {
		t.Fatalf("expected host error, got %v", err)
	}
}

func TestValidateEmailEnabledRequiresFrom(t *testing.T) {
	s := settingtypes.DefaultApplicationSettings()
	s.Email.Enabled = true
	s.Email.Host = "smtp.example.com"
	s.Email.From = "not-an-email"
	if err := s.Validate(); err == nil || !strings.Contains(err.Error(), "from") {
		t.Fatalf("expected from error, got %v", err)
	}
}

func TestValidateTLSInvalid(t *testing.T) {
	s := settingtypes.DefaultApplicationSettings()
	s.Email.TLS = "weird"
	if err := s.Validate(); err == nil || !strings.Contains(err.Error(), "tls") {
		t.Fatalf("expected tls error, got %v", err)
	}
}

func TestMergePatchPassword(t *testing.T) {
	base := settingtypes.DefaultApplicationSettings()
	base.Email.Password = "secret"
	pw := ""
	patch := &settingtypes.SettingsPatchRequest{
		Email: &settingtypes.EmailSMTPPatch{Password: &pw},
	}
	out, err := settingtypes.MergePatch(base, patch)
	if err != nil {
		t.Fatalf("MergePatch: %v", err)
	}
	if out.Email.Password != "" {
		t.Fatalf("password = %q, want empty", out.Email.Password)
	}
}

func TestMergePatchOmitPasswordKeeps(t *testing.T) {
	base := settingtypes.DefaultApplicationSettings()
	base.Email.Password = "keep"
	host := "smtp.example.com"
	patch := &settingtypes.SettingsPatchRequest{
		Email: &settingtypes.EmailSMTPPatch{Host: &host},
	}
	out, err := settingtypes.MergePatch(base, patch)
	if err != nil {
		t.Fatalf("MergePatch: %v", err)
	}
	if out.Email.Password != "keep" {
		t.Fatalf("password = %q, want keep", out.Email.Password)
	}
}

func TestParsePayloadEmpty(t *testing.T) {
	s, err := settingtypes.ParsePayload(nil)
	if err != nil {
		t.Fatalf("ParsePayload: %v", err)
	}
	if s.Email.Port != 587 {
		t.Fatalf("port = %d", s.Email.Port)
	}
}
