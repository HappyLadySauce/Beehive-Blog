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
