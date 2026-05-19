package settings

import (
	"testing"

	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

func TestProviderDefaultsToEmptySnapshot(t *testing.T) {
	p := NewProvider()
	if p.CachedRevision() != 0 {
		t.Fatalf("revision = %d, want 0", p.CachedRevision())
	}
	cur := p.Current()
	if cur.Email.Port != settingtypes.DefaultApplicationSettings().Email.Port {
		t.Fatalf("port = %d", cur.Email.Port)
	}
}

func TestProviderReplaceUpdatesSnapshot(t *testing.T) {
	p := NewProvider()
	next := settingtypes.DefaultApplicationSettings()
	next.Email.Host = "smtp.example.com"

	p.Replace(next, 7)

	if p.CachedRevision() != 7 {
		t.Fatalf("revision = %d, want 7", p.CachedRevision())
	}
	if got := p.Current().Email.Host; got != "smtp.example.com" {
		t.Fatalf("host = %q", got)
	}
}
