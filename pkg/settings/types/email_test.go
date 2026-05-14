package types

import (
	"strings"
	"testing"
)

func TestDefaultApplicationSettings(t *testing.T) {
	s := DefaultApplicationSettings()
	if s.Email.Enabled {
		t.Fatalf("Email.Enabled: got true, want false")
	}
	if s.Email.Port != 587 {
		t.Fatalf("Email.Port: got %d, want 587", s.Email.Port)
	}
	if s.Email.TLS != EmailTLSStartTLS {
		t.Fatalf("Email.TLS: got %q, want %q", s.Email.TLS, EmailTLSStartTLS)
	}
	if s.Email.Host != "" || s.Email.Username != "" || s.Email.Password != "" || s.Email.From != "" || s.Email.FromName != "" {
		t.Fatalf("Email string fields should be empty in defaults: %+v", s.Email)
	}
	if err := s.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestEmailSMTPSettingsValidateForSend(t *testing.T) {
	enabled := true
	valid := EmailSMTPSettings{
		Enabled:  enabled,
		Host:     "smtp.example.com",
		Port:     587,
		From:     "noreply@example.com",
		FromName: "Beehive",
		TLS:      EmailTLSStartTLS,
	}
	tests := []struct {
		name    string
		e       EmailSMTPSettings
		wantErr string
	}{
		{
			name: "disabled minimal",
			e: EmailSMTPSettings{
				Enabled: false,
				TLS:     EmailTLSStartTLS,
			},
		},
		{
			name: "disabled tls none",
			e: EmailSMTPSettings{
				Enabled: false,
				TLS:     EmailTLSNone,
			},
		},
		{
			name: "disabled tls direct",
			e: EmailSMTPSettings{
				Enabled: false,
				TLS:     EmailTLSDirect,
			},
		},
		{
			name: "tls trimmed and lowercased",
			e: EmailSMTPSettings{
				Enabled: false,
				TLS:     "  STARTTLS  ",
			},
		},
		{
			name:    "invalid tls even when disabled",
			e:       EmailSMTPSettings{Enabled: false, TLS: "weird"},
			wantErr: "email.tls",
		},
		{
			name:    "empty tls without normalize",
			e:       EmailSMTPSettings{Enabled: false, TLS: ""},
			wantErr: "email.tls",
		},
		{
			name:    "enabled missing host",
			e:       EmailSMTPSettings{Enabled: true, Port: 587, From: "a@b.co", TLS: EmailTLSStartTLS},
			wantErr: "email.host",
		},
		{
			name: "enabled port too low",
			e: EmailSMTPSettings{
				Enabled: true, Host: "h", Port: 0, From: "a@b.co", TLS: EmailTLSStartTLS,
			},
			wantErr: "email.port",
		},
		{
			name: "enabled port too high",
			e: EmailSMTPSettings{
				Enabled: true, Host: "h", Port: 65536, From: "a@b.co", TLS: EmailTLSStartTLS,
			},
			wantErr: "email.port",
		},
		{
			name: "enabled missing from",
			e: EmailSMTPSettings{
				Enabled: true, Host: "h", Port: 25, TLS: EmailTLSStartTLS,
			},
			wantErr: "email.from",
		},
		{
			name: "enabled invalid from",
			e: EmailSMTPSettings{
				Enabled: true, Host: "h", Port: 25, From: "not-an-email", TLS: EmailTLSStartTLS,
			},
			wantErr: "email.from",
		},
		{
			name: "enabled angle address",
			e: EmailSMTPSettings{
				Enabled: true, Host: "h", Port: 25, From: `"Site" <site@example.com>`, TLS: EmailTLSStartTLS,
			},
		},
		{
			name:    "valid",
			e:       valid,
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.e.ValidateForSend()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateForSend: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ValidateForSend: got nil, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("ValidateForSend: got %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}
