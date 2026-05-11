package session_test

import (
	"testing"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/session"
)

func TestHashRefreshTokenReturnsStableDigestNotPlaintext(t *testing.T) {
	token := "refresh-token-value"
	hash := session.HashRefreshToken(token)
	if hash == token {
		t.Fatalf("hash must not equal plaintext token")
	}
	if len(hash) != 64 {
		t.Fatalf("hash length = %d, want 64", len(hash))
	}
	if got := session.HashRefreshToken(token); got != hash {
		t.Fatalf("hash is not stable: %q != %q", got, hash)
	}
}

func TestNewJTIProducesUniqueHexIDs(t *testing.T) {
	first, err := session.NewJTI()
	if err != nil {
		t.Fatalf("NewJTI() error = %v", err)
	}
	second, err := session.NewJTI()
	if err != nil {
		t.Fatalf("NewJTI() error = %v", err)
	}
	if first == second {
		t.Fatalf("NewJTI() returned duplicate value %q", first)
	}
	if len(first) != 64 || len(second) != 64 {
		t.Fatalf("unexpected jti lengths: %d, %d", len(first), len(second))
	}
}

func TestTruncateStringTruncatesAtLimit(t *testing.T) {
	val := "abcdefghij"
	if got := session.TruncateString(val, 5); got != "abcde" {
		t.Fatalf("TruncateString(%q, 5) = %q, want abcde", val, got)
	}
}

func TestTruncateStringPreservesShortStrings(t *testing.T) {
	val := "abc"
	if got := session.TruncateString(val, 10); got != "abc" {
		t.Fatalf("TruncateString(%q, 10) = %q, want abc", val, got)
	}
}

func TestTruncateStringPreservesExactLength(t *testing.T) {
	val := "abcde"
	if got := session.TruncateString(val, 5); got != "abcde" {
		t.Fatalf("TruncateString(%q, 5) = %q, want abcde", val, got)
	}
}

func TestRevokeSessionIdempotentOnZeroIDs(t *testing.T) {
	if err := session.RevokeSession(nil, 0, 42, "logout"); err != nil {
		t.Fatalf("RevokeSession(0, 42) error = %v, want nil", err)
	}
	if err := session.RevokeSession(nil, 7, 0, "logout"); err != nil {
		t.Fatalf("RevokeSession(7, 0) error = %v, want nil", err)
	}
}

func TestRevokeUserSessionsIdempotentOnZeroID(t *testing.T) {
	if err := session.RevokeUserSessions(nil, 0, "reason"); err != nil {
		t.Fatalf("RevokeUserSessions(0) error = %v, want nil", err)
	}
}
