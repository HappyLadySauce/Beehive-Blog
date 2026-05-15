package attachments_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	routeattachments "github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/attachments"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	pkgattachment "github.com/HappyLadySauce/Beehive-Blog/pkg/attachment"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/config"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

func TestNewAttachmentsControllerValidation(t *testing.T) {
	t.Run("nil service context", func(t *testing.T) {
		_, err := routeattachments.NewAttachmentsController(nil)
		if err == nil || !strings.Contains(err.Error(), "service context is nil") {
			t.Fatalf("NewAttachmentsController: %v", err)
		}
	})
	t.Run("nil config", func(t *testing.T) {
		_, err := routeattachments.NewAttachmentsController(&svc.ServiceContext{DB: newGormTestDB(t)})
		if err == nil || !strings.Contains(err.Error(), "config is nil") {
			t.Fatalf("NewAttachmentsController: %v", err)
		}
	})
	t.Run("nil database", func(t *testing.T) {
		_, err := routeattachments.NewAttachmentsController(&svc.ServiceContext{
			Config: &config.Config{},
		})
		if err == nil || !strings.Contains(err.Error(), "database handle is nil") {
			t.Fatalf("NewAttachmentsController: %v", err)
		}
	})
}

func TestNewAttachmentsControllerSuccess(t *testing.T) {
	h, err := routeattachments.NewAttachmentsController(&svc.ServiceContext{
		DB:     newGormTestDB(t),
		Config: &config.Config{},
	})
	if err != nil {
		t.Fatalf("NewAttachmentsController: %v", err)
	}
	if h == nil {
		t.Fatal("controller is nil")
	}
}

func TestListRequiresAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := mustNewAttachmentsController(t)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachments", nil)
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: "member"})
	h.List(ctx)
	assertEnvelopeCode(t, rec, http.StatusForbidden)
}

func TestListInvalidPurpose(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := mustNewAttachmentsController(t)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachments?purpose=unknown-purpose", nil)
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
	h.List(ctx)
	assertEnvelopeCode(t, rec, http.StatusBadRequest)
}

func TestListInvalidStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := mustNewAttachmentsController(t)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachments?status=bogus", nil)
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
	h.List(ctx)
	assertEnvelopeCode(t, rec, http.StatusBadRequest)
}

func TestListInvalidOwnerUserIDQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := mustNewAttachmentsController(t)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachments?owner_user_id=notnum", nil)
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
	h.List(ctx)
	assertEnvelopeCode(t, rec, http.StatusBadRequest)
}

func TestListInvalidCursor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := mustNewAttachmentsController(t)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachments?cursor=bad", nil)
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
	h.List(ctx)
	assertEnvelopeCode(t, rec, http.StatusBadRequest)
}

func TestListInvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := mustNewAttachmentsController(t)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachments?limit=101", nil)
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
	h.List(ctx)
	assertEnvelopeCode(t, rec, http.StatusBadRequest)
}

func TestCreateCategoryValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := mustNewAttachmentsController(t)
	t.Run("empty name", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		body := `{"name":"   ","slug":"s"}`
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/attachment/categories", strings.NewReader(body))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
		h.CreateCategory(ctx)
		assertEnvelopeCode(t, rec, http.StatusBadRequest)
	})
	t.Run("invalid status", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		body := `{"name":"n","slug":"s","status":"unknown"}`
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/attachment/categories", strings.NewReader(body))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
		h.CreateCategory(ctx)
		assertEnvelopeCode(t, rec, http.StatusBadRequest)
	})
}

func TestGetAttachmentInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := mustNewAttachmentsController(t)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachments/abc", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "abc"}}
	h.GetAttachment(ctx)
	assertEnvelopeCode(t, rec, http.StatusBadRequest)
}

func TestGetAttachmentInvalidBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	issuer := testAttachmentJWT(t)
	h, err := routeattachments.NewAttachmentsController(&svc.ServiceContext{
		DB:     newGormTestDB(t),
		Config: &config.Config{},
		Token:  issuer,
	})
	if err != nil {
		t.Fatalf("NewAttachmentsController: %v", err)
	}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachments/1", nil)
	ctx.Request.Header.Set("Authorization", "Bearer not-a-jwt")
	ctx.Params = gin.Params{{Key: "id", Value: "1"}}
	h.GetAttachment(ctx)
	assertEnvelopeCode(t, rec, http.StatusUnauthorized)
}

func TestUploadLocalInvalidOwnerUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := mustNewAttachmentsController(t)
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("owner_user_id", "0")
	fw, err := w.CreateFormFile("file", "a.txt")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := fw.Write([]byte("hi")); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/attachments", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	ctx.Request = req
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
	h.UploadLocal(ctx)
	assertEnvelopeCode(t, rec, http.StatusBadRequest)
}

func mustNewAttachmentsController(t *testing.T) *routeattachments.AttachmentsController {
	t.Helper()
	h, err := routeattachments.NewAttachmentsController(&svc.ServiceContext{
		DB:     newGormTestDB(t),
		Config: &config.Config{},
	})
	if err != nil {
		t.Fatalf("NewAttachmentsController: %v", err)
	}
	return h
}

func newGormTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return db
}

func assertEnvelopeCode(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("HTTP status = %d, want %d, body=%s", rec.Code, want, rec.Body.String())
	}
	var env common.BaseResponse
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v body=%s", err, rec.Body.String())
	}
	if env.Code != want {
		t.Fatalf("envelope code = %d, want %d", env.Code, want)
	}
}

func testAttachmentJWT(t *testing.T) *jwt.Issuer {
	t.Helper()
	issuer, err := jwt.NewIssuer(&options.JWTOptions{
		Issuer:     "beehive-blog-attachments-test",
		Secret:     "0123456789abcdef0123456789abcdef",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	})
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}
	return issuer
}
