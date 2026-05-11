package jwt_test

import (
	"testing"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

func newTestIssuer(t *testing.T) *jwt.Issuer {
	t.Helper()
	issuer, err := jwt.NewIssuer(&options.JWTOptions{
		Issuer:     "beehive-blog-test",
		Secret:     "0123456789abcdef0123456789abcdef",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	})
	if err != nil {
		t.Fatalf("NewIssuer() error = %v", err)
	}
	return issuer
}

func TestIssueSessionPairBindsSessionAndRefreshJTI(t *testing.T) {
	issuer := newTestIssuer(t)
	pair, err := issuer.IssueSessionPair(42, "member", 99, "refresh-jti")
	if err != nil {
		t.Fatalf("IssueSessionPair() error = %v", err)
	}

	accessClaims, err := issuer.ParseAccess(pair.Access.Token)
	if err != nil {
		t.Fatalf("ParseAccess() error = %v", err)
	}
	if accessClaims.SID != 99 {
		t.Fatalf("access sid = %d, want 99", accessClaims.SID)
	}

	refreshClaims, err := issuer.ParseRefresh(pair.Refresh.Token)
	if err != nil {
		t.Fatalf("ParseRefresh() error = %v", err)
	}
	if refreshClaims.SID != 99 {
		t.Fatalf("refresh sid = %d, want 99", refreshClaims.SID)
	}
	if refreshClaims.ID != "refresh-jti" {
		t.Fatalf("refresh jti = %q, want refresh-jti", refreshClaims.ID)
	}
}
