package types

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestApplicationSettingsNormalize(t *testing.T) {
	t.Run("fills default port and tls", func(t *testing.T) {
		s := ApplicationSettings{
			Email: EmailSMTPSettings{Port: 0, TLS: ""},
		}
		s.Normalize()
		if s.Email.Port != 587 {
			t.Fatalf("Port: got %d, want 587", s.Email.Port)
		}
		if s.Email.TLS != EmailTLSStartTLS {
			t.Fatalf("TLS: got %q, want %q", s.Email.TLS, EmailTLSStartTLS)
		}
	})
	t.Run("preserves positive port", func(t *testing.T) {
		s := ApplicationSettings{
			Email: EmailSMTPSettings{Port: 25, TLS: EmailTLSNone},
		}
		s.Normalize()
		if s.Email.Port != 25 {
			t.Fatalf("Port: got %d, want 25", s.Email.Port)
		}
	})
}

func TestApplicationSettingsValidate(t *testing.T) {
	s := ApplicationSettings{
		Email: EmailSMTPSettings{
			Enabled: false,
			Port:    0,
			TLS:     "",
		},
	}
	if err := s.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if s.Email.Port != 587 || s.Email.TLS != EmailTLSStartTLS {
		t.Fatalf("Validate should normalize in place: %+v", s.Email)
	}
}

func TestMergePatch(t *testing.T) {
	base := DefaultApplicationSettings()

	t.Run("nil patch", func(t *testing.T) {
		_, err := MergePatch(base, nil)
		if err == nil || !strings.Contains(err.Error(), "patch is nil") {
			t.Fatalf("MergePatch: got err=%v", err)
		}
	})

	t.Run("nil email leaves base unchanged", func(t *testing.T) {
		out, err := MergePatch(base, &SettingsPatchRequest{Email: nil})
		if err != nil {
			t.Fatalf("MergePatch: %v", err)
		}
		if out != base {
			t.Fatalf("expected same values as base; got %+v want %+v", out, base)
		}
	})

	t.Run("partial patch rejected when result invalid", func(t *testing.T) {
		on := true
		host := "smtp.example.org"
		_, err := MergePatch(base, &SettingsPatchRequest{
			Email: &EmailSMTPPatch{Enabled: &on, Host: &host},
		})
		if err == nil {
			t.Fatal("expected validation error when enabled without from")
		}
	})

	t.Run("successful merge", func(t *testing.T) {
		on := true
		host := "smtp.example.org"
		port := 465
		from := "mailer@example.org"
		tls := EmailTLSDirect
		user := "u"
		pass := "p"
		name := "Alerts"
		out, err := MergePatch(base, &SettingsPatchRequest{
			Email: &EmailSMTPPatch{
				Enabled:  &on,
				Host:     &host,
				Port:     &port,
				Username: &user,
				Password: &pass,
				From:     &from,
				FromName: &name,
				TLS:      &tls,
			},
		})
		if err != nil {
			t.Fatalf("MergePatch: %v", err)
		}
		if !out.Email.Enabled || out.Email.Host != host || out.Email.Port != port ||
			out.Email.Username != user || out.Email.Password != pass || out.Email.From != from ||
			out.Email.FromName != name || out.Email.TLS != tls {
			t.Fatalf("unexpected merged settings: %+v", out.Email)
		}
	})
}

func TestParsePayload(t *testing.T) {
	t.Run("empty uses defaults", func(t *testing.T) {
		s, err := ParsePayload(nil)
		if err != nil {
			t.Fatalf("ParsePayload: %v", err)
		}
		want := DefaultApplicationSettings()
		if s != want {
			t.Fatalf("got %+v, want %+v", s, want)
		}
	})
	t.Run("null uses defaults", func(t *testing.T) {
		s, err := ParsePayload([]byte("null"))
		if err != nil {
			t.Fatalf("ParsePayload: %v", err)
		}
		if s != DefaultApplicationSettings() {
			t.Fatalf("got %+v", s)
		}
	})
	t.Run("invalid json", func(t *testing.T) {
		_, err := ParsePayload([]byte("{"))
		if err == nil || !strings.Contains(err.Error(), "decode settings payload") {
			t.Fatalf("ParsePayload: got err=%v", err)
		}
	})
	t.Run("valid minimal json", func(t *testing.T) {
		raw := []byte(`{"email":{"enabled":false}}`)
		s, err := ParsePayload(raw)
		if err != nil {
			t.Fatalf("ParsePayload: %v", err)
		}
		if s.Email.Enabled {
			t.Fatalf("Enabled: got true")
		}
		if s.Email.Port != 587 || s.Email.TLS != EmailTLSStartTLS {
			t.Fatalf("expected defaults after normalize: %+v", s.Email)
		}
	})
	t.Run("reject invalid tls in payload", func(t *testing.T) {
		raw := []byte(`{"email":{"enabled":false,"tls":"bogus"}}`)
		_, err := ParsePayload(raw)
		if err == nil || !strings.Contains(err.Error(), "email.tls") {
			t.Fatalf("ParsePayload: got err=%v", err)
		}
	})
}

func TestMarshalPayload(t *testing.T) {
	s := DefaultApplicationSettings()
	raw, err := MarshalPayload(s)
	if err != nil {
		t.Fatalf("MarshalPayload: %v", err)
	}
	if !bytes.Contains(raw, []byte(`"enabled":false`)) {
		t.Fatalf("unexpected JSON: %s", string(raw))
	}
	var decoded ApplicationSettings
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	decoded.Normalize()
	s.Normalize()
	if decoded != s {
		t.Fatalf("round-trip mismatch: got %+v want %+v", decoded, s)
	}
}
