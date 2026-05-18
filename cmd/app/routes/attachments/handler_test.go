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

func TestInitRegistersReferenceRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("Init panicked while registering routes: %v", recovered)
		}
	}()
	if err := routeattachments.Init(&svc.ServiceContext{
		DB:     newGormTestDB(t),
		Config: &config.Config{},
	}); err != nil {
		t.Fatalf("Init: %v", err)
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

func TestListInvalidCategoryMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := mustNewAttachmentsController(t)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachments?category_mode=unknown", nil)
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
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachments?limit=201", nil)
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
	h.List(ctx)
	assertEnvelopeCode(t, rec, http.StatusBadRequest)
}

func TestListInvalidPageSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := mustNewAttachmentsController(t)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachments?page=1&page_size=201", nil)
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
	h.List(ctx)
	assertEnvelopeCode(t, rec, http.StatusBadRequest)
}

func TestListOffsetPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock := mustNewAttachmentsControllerWithMock(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT count\(\*\)`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(25)))
	mock.ExpectQuery(`SELECT .* FROM "attachment"."attachments"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "purpose", "filename", "mime_type", "size", "storage_mount_id", "object_key", "storage_metadata", "access_scope", "upload_status", "status", "created_at", "updated_at", "deleted_at",
		}).AddRow(int64(80), "content", "note.md", "text/markdown", int64(128), int64(10), "content/note.md", []byte(`{}`), "private", "ready", "active", now, now, nil))

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachments?page=2&page_size=20", nil)
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
	h.List(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope struct {
		Code int `json:"code"`
		Data struct {
			Total    int64 `json:"total"`
			Page     int   `json:"page"`
			PageSize int   `json:"page_size"`
			Items    []struct {
				ID int64 `json:"id"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, rec.Body.String())
	}
	if envelope.Code != http.StatusOK {
		t.Fatalf("envelope code = %d", envelope.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
	if envelope.Data.Total != 25 || envelope.Data.Page != 2 || envelope.Data.PageSize != 20 {
		t.Fatalf("pagination metadata: %+v", envelope.Data)
	}
	if len(envelope.Data.Items) != 1 || envelope.Data.Items[0].ID != 80 {
		t.Fatalf("items: %+v", envelope.Data.Items)
	}
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

func TestUploadBatchRejectsInvalidMultipart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := mustNewAttachmentsController(t)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/batch", strings.NewReader("not-a-valid-multipart-body"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=broken")
	ctx.Request = req
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})

	h.UploadBatch(ctx)

	assertEnvelopeCode(t, rec, http.StatusBadRequest)
}

func TestGetReferencesReturnsUserAvatarReference(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock := mustNewAttachmentsControllerWithMock(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .* FROM "identity"."users"`).
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "nickname", "avatar_attachment_id", "role", "status", "created_at", "updated_at", "deleted_at",
		}).AddRow(int64(1), "admin", "Admin", int64(99), "admin", "active", now, now, nil))

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachments/99/references", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "99"}}
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
	h.GetReferences(ctx)

	body := rec.Body.String()
	assertEnvelopeCode(t, rec, http.StatusOK)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
	if !strings.Contains(body, `"source_type":"user"`) || !strings.Contains(body, `"relation":"avatar"`) {
		t.Fatalf("reference response missing user avatar reference: %s", body)
	}
}

func TestGetReferencesRequiresAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := mustNewAttachmentsController(t)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachments/99/references", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "99"}}
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: "member"})
	h.GetReferences(ctx)
	assertEnvelopeCode(t, rec, http.StatusForbidden)
}

func TestDeleteRejectsReferencedAttachment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock := mustNewAttachmentsControllerWithMock(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .* FROM "attachment"."attachments"`).
		WithArgs(int64(99), 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "purpose", "filename", "mime_type", "size", "storage_mount_id", "object_key", "storage_metadata", "access_scope", "upload_status", "status", "created_at", "updated_at", "deleted_at",
		}).AddRow(int64(99), "content", "note.md", "text/markdown", int64(128), int64(10), "content/note.md", []byte(`{}`), "private", "ready", "active", now, now, nil))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "identity"."users"`).
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/attachments/99", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "99"}}
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
	h.Delete(ctx)

	assertEnvelopeCode(t, rec, http.StatusConflict)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestDeleteForceClearsUserAvatarReferences(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock := mustNewAttachmentsControllerWithMock(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .* FROM "attachment"."attachments"`).
		WithArgs(int64(99), 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "purpose", "filename", "mime_type", "size", "storage_mount_id", "object_key", "storage_metadata", "access_scope", "upload_status", "status", "created_at", "updated_at", "deleted_at",
		}).AddRow(int64(99), "content", "note.md", "text/markdown", int64(128), int64(10), "content/note.md", []byte(`{}`), "private", "ready", "active", now, now, nil))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "identity"."users"`).
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "identity"\."users" SET .*"avatar_attachment_id"=\$1.*"updated_at"=\$2.*WHERE avatar_attachment_id = \$3 AND "users"\."deleted_at" IS NULL`).
		WithArgs(nil, sqlmock.AnyArg(), int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE "attachment"\."attachments" SET "deleted_at"=\$1 WHERE id = \$2 AND "attachments"\."deleted_at" IS NULL`).
		WithArgs(sqlmock.AnyArg(), int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/attachments/99?force=true", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "99"}}
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
	h.Delete(ctx)

	assertEnvelopeCode(t, rec, http.StatusOK)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
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

func mustNewAttachmentsControllerWithMock(t *testing.T) (*routeattachments.AttachmentsController, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	h, err := routeattachments.NewAttachmentsController(&svc.ServiceContext{
		DB:     db,
		Config: &config.Config{},
	})
	if err != nil {
		t.Fatalf("NewAttachmentsController: %v", err)
	}
	return h, mock
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
