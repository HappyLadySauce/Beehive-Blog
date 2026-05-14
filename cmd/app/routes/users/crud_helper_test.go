package users_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	routeusers "github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/users"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

type crudEnvelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"message"`
	Data json.RawMessage `json:"data"`
}

func newCrudTestController(t *testing.T) (*routeusers.UsersController, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	issuer, err := jwt.NewIssuer(&options.JWTOptions{
		Secret:     "0123456789abcdef0123456789abcdef",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
		Issuer:     "beehive-blog-crud-test",
	})
	if err != nil {
		t.Fatalf("jwt.NewIssuer: %v", err)
	}
	return routeusers.NewUsersController(&svc.ServiceContext{
		DB:    gdb,
		Token: issuer,
	}), mock
}

func testCrudContext(method, path string, body *bytes.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	if body == nil {
		body = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	return ctx, rec
}

func testCrudContextWithID(method, path string, body *bytes.Reader, id string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	if body == nil {
		body = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "id", Value: id}}
	return ctx, rec
}

func decodeCrudEnvelope(t *testing.T, rec *httptest.ResponseRecorder) crudEnvelope {
	t.Helper()
	var env crudEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return env
}

func assertCrudError(t *testing.T, rec *httptest.ResponseRecorder, env crudEnvelope, status int, msg string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("HTTP = %d, want %d", rec.Code, status)
	}
	if env.Code != status {
		t.Fatalf("code = %d, want %d", env.Code, status)
	}
	if env.Msg != msg {
		t.Fatalf("message = %q, want %q", env.Msg, msg)
	}
}
