package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jwt5 "github.com/golang-jwt/jwt/v5"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
)

func TestSessionReturnsVerifiedAccessClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{
		UID:  42,
		Role: "admin",
		SID:  7,
		RegisteredClaims: jwt5.RegisteredClaims{
			ExpiresAt: jwt5.NewNumericDate(time.Unix(4_102_444_800, 0)),
		},
	})

	controller := NewAuthController(nil)
	controller.Session(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	for _, want := range []string{`"uid":42`, `"role":"admin"`, `"exp":4102444800`, `"sid":7`} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %s, want field %s", body, want)
		}
	}
}

func TestSessionRejectsMissingVerifiedClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)

	controller := NewAuthController(nil)
	controller.Session(ctx)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
