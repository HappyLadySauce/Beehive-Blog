package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
)

func TestGetClaimsMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	if got := GetClaims(c); got != nil {
		t.Fatalf("GetClaims() = %v, want nil", got)
	}
}

func TestGetClaimsWrongType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(jwt.ClaimsKey, "not-claims")
	if got := GetClaims(c); got != nil {
		t.Fatalf("GetClaims() = %v, want nil for wrong type", got)
	}
}

func TestGetClaimsReturnsInjectedClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	want := &jwt.Claims{UID: 42, Role: "admin"}
	c.Set(jwt.ClaimsKey, want)
	if got := GetClaims(c); got != want {
		t.Fatalf("GetClaims() = %p, want %p", got, want)
	}
}
