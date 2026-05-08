package session

import "testing"

func TestHashRefreshTokenReturnsStableDigestNotPlaintext(t *testing.T) {
	token := "refresh-token-value"
	hash := HashRefreshToken(token)
	if hash == token {
		t.Fatalf("hash must not equal plaintext token")
	}
	if len(hash) != 64 {
		t.Fatalf("hash length = %d, want 64", len(hash))
	}
	if got := HashRefreshToken(token); got != hash {
		t.Fatalf("hash is not stable: %q != %q", got, hash)
	}
}

func TestNewJTIProducesUniqueHexIDs(t *testing.T) {
	first, err := NewJTI()
	if err != nil {
		t.Fatalf("NewJTI() error = %v", err)
	}
	second, err := NewJTI()
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
