package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

func testOptionalAuthIssuer(t *testing.T) *jwt.Issuer {
	t.Helper()
	issuer, err := jwt.NewIssuer(&options.JWTOptions{
		Issuer:     "beehive-blog-optional-auth-test",
		Secret:     "0123456789abcdef0123456789abcdef",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	})
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}
	return issuer
}

func serveOptionalAuthProbe(t *testing.T, issuer *jwt.Issuer, authHeader string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	engine := gin.New()
	engine.Use(OptionalAuthMiddleware(&svc.ServiceContext{Token: issuer}))
	engine.GET("/probe", func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			c.JSON(http.StatusOK, gin.H{"admin": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{"admin": claims.Role == "admin"})
	})
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	engine.ServeHTTP(rec, req)
	return rec
}

func TestOptionalAuthMiddlewareNoHeader(t *testing.T) {
	rec := serveOptionalAuthProbe(t, testOptionalAuthIssuer(t), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"admin":false`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestOptionalAuthMiddlewareValidBearer(t *testing.T) {
	issuer := testOptionalAuthIssuer(t)
	pair, err := issuer.IssuePair(10, "admin")
	if err != nil {
		t.Fatalf("IssuePair: %v", err)
	}
	rec := serveOptionalAuthProbe(t, issuer, "Bearer "+pair.Access.Token)
	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"admin":true`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestOptionalAuthMiddlewareInvalidBearer(t *testing.T) {
	rec := serveOptionalAuthProbe(t, testOptionalAuthIssuer(t), "Bearer not-a-jwt")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("HTTP = %d, want 401", rec.Code)
	}
}

func TestOptionalAuthMiddlewareMalformedHeader(t *testing.T) {
	rec := serveOptionalAuthProbe(t, testOptionalAuthIssuer(t), "Basic abc")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("HTTP = %d, want 401", rec.Code)
	}
}
