package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
)

func TestRequireRoleAllowsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Set(jwt.ClaimsKey, &jwt.Claims{Role: "admin"})
	RequireRole("admin")(c)
	if c.IsAborted() {
		t.Fatal("expected not aborted for admin")
	}
	if rec.Code == http.StatusForbidden {
		t.Fatal("unexpected 403 for admin")
	}
}

func TestRequireRoleRejectsMember(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Set(jwt.ClaimsKey, &jwt.Claims{Role: "member"})
	RequireRole("admin")(c)
	if !c.IsAborted() {
		t.Fatal("expected aborted")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("code = %d, want 403", rec.Code)
	}
}
