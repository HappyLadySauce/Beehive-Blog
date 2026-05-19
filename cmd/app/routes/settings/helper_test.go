package settings

import (
	"testing"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

func TestToResponsePasswordAndSecretFlags(t *testing.T) {
	s := settingtypes.DefaultApplicationSettings()
	s.Email.Password = "secret"
	s.GithubOAuth2.ClientSecret = "oauth-secret"

	got := toResponse(s, 11)
	if !got.Email.PasswordSet {
		t.Fatal("PasswordSet = false, want true")
	}
	if !got.GithubOAuth2.ClientSecretSet {
		t.Fatal("ClientSecretSet = false, want true")
	}
	if got.Revision != 11 {
		t.Fatalf("revision = %d, want 11", got.Revision)
	}
}

func TestToResponseEmptySecretsNotSet(t *testing.T) {
	s := settingtypes.DefaultApplicationSettings()
	got := toResponse(s, 3)
	if got.Email.PasswordSet {
		t.Fatal("PasswordSet = true, want false for empty password")
	}
	if got.GithubOAuth2.ClientSecretSet {
		t.Fatal("ClientSecretSet = true, want false for empty secret")
	}
}

func TestPatchFromV1Nil(t *testing.T) {
	if patchFromV1(nil) != nil {
		t.Fatal("patchFromV1(nil) should be nil")
	}
}

func TestPatchFromV1MapsFields(t *testing.T) {
	enabled := true
	host := "smtp.example.com"
	port := 587
	p := &v1.EmailSMTPPatchJSON{
		Enabled: &enabled,
		Host:    &host,
		Port:    &port,
	}
	got := patchFromV1(p)
	if got == nil || got.Host == nil || *got.Host != host {
		t.Fatalf("patchFromV1() = %+v", got)
	}
}

func TestPatchGithubFromV1Nil(t *testing.T) {
	if patchGithubFromV1(nil) != nil {
		t.Fatal("patchGithubFromV1(nil) should be nil")
	}
}

func TestPatchGithubFromV1MapsClientID(t *testing.T) {
	id := "client-id"
	got := patchGithubFromV1(&v1.GithubOAuth2PatchJSON{ClientID: &id})
	if got == nil || got.ClientID == nil || *got.ClientID != id {
		t.Fatalf("patchGithubFromV1() = %+v", got)
	}
}
