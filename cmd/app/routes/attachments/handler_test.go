package attachments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	pkgattachment "github.com/HappyLadySauce/Beehive-Blog/pkg/attachment"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/attachment/driver"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/config"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

func TestNewAttachmentsControllerValidation(t *testing.T) {
	t.Run("nil service context", func(t *testing.T) {
		_, err := NewAttachmentsController(nil)
		if err == nil || !strings.Contains(err.Error(), "service context is nil") {
			t.Fatalf("NewAttachmentsController: %v", err)
		}
	})
	t.Run("nil config", func(t *testing.T) {
		_, err := NewAttachmentsController(&svc.ServiceContext{DB: newGormTestDB(t)})
		if err == nil || !strings.Contains(err.Error(), "config is nil") {
			t.Fatalf("NewAttachmentsController: %v", err)
		}
	})
	t.Run("nil database", func(t *testing.T) {
		_, err := NewAttachmentsController(&svc.ServiceContext{
			Config: &config.Config{},
		})
		if err == nil || !strings.Contains(err.Error(), "database handle is nil") {
			t.Fatalf("NewAttachmentsController: %v", err)
		}
	})
}

func TestNewAttachmentsControllerSuccess(t *testing.T) {
	h, err := NewAttachmentsController(&svc.ServiceContext{
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
	if err := Init(&svc.ServiceContext{
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
	svcCtx := &svc.ServiceContext{
		DB:     newGormTestDB(t),
		Config: &config.Config{},
		Token:  issuer,
	}
	h, err := NewAttachmentsController(svcCtx)
	if err != nil {
		t.Fatalf("NewAttachmentsController: %v", err)
	}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachments/1", nil)
	ctx.Request.Header.Set("Authorization", "Bearer not-a-jwt")
	ctx.Params = gin.Params{{Key: "id", Value: "1"}}
	middleware.OptionalAuthMiddleware(svcCtx)(ctx)
	if !ctx.IsAborted() {
		h.GetAttachment(ctx)
	}
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

func TestUploadLocalCleansStorageObjectWhenDatabaseTransactionFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, fakeDrv := mustNewAttachmentsControllerWithFakeDriver(t)
	mountID := int64(10)
	now := time.Now()

	mock.ExpectQuery(`SELECT .* FROM "attachment"\."storage_mounts"`).
		WithArgs(mountID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "driver_name", "mount_path", "name", "config", "order_index", "is_default", "disabled", "status", "created_at", "updated_at", "deleted_at",
		}).AddRow(mountID, "fake", "/", "fake", []byte(`{}`), 0, true, false, "ready", now, now, nil))
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "attachment"\."attachments"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(99)))
	mock.ExpectExec(`DELETE FROM "attachment"\."attachment_categories" WHERE attachment_id = \$1`).
		WithArgs(int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "attachment"\."categories"`).
		WithArgs(int64(404), pkgattachment.CategoryStatusActive).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectRollback()

	_, err := h.uploadLocal(context.Background(), pkgattachment.Actor{Role: pkgattachment.RoleAdmin}, pkgattachment.LocalUploadInput{
		Purpose:        pkgattachment.PurposeContent,
		Filename:       "leak.txt",
		MimeType:       "text/plain",
		Size:           int64(len("hi")),
		Reader:         strings.NewReader("hi"),
		AccessScope:    pkgattachment.AccessPrivate,
		CategoryIDs:    []int64{404},
		StorageMountID: &mountID,
	})
	if err == nil {
		t.Fatal("uploadLocal should fail when category binding validation fails")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
	if len(fakeDrv.deletedKeys) != 1 {
		t.Fatalf("deleted keys = %v, want one cleanup", fakeDrv.deletedKeys)
	}
	if fakeDrv.deletedKeys[0] != fakeDrv.savedKey {
		t.Fatalf("deleted key = %q, want saved key %q", fakeDrv.deletedKeys[0], fakeDrv.savedKey)
	}
}

func TestUploadBatchRejectsTooManyFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := mustNewAttachmentsController(t)
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for i := 0; i < pkgattachment.MaxBatchUploadFiles+1; i++ {
		part, err := w.CreateFormFile("files", fmt.Sprintf("file%d.txt", i))
		if err != nil {
			t.Fatalf("CreateFormFile: %v", err)
		}
		if _, err := part.Write([]byte("x")); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/attachments/batch", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	ctx.Request = req
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
	h.UploadBatch(ctx)
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

func mustNewAttachmentsController(t *testing.T) *AttachmentsController {
	t.Helper()
	h, err := NewAttachmentsController(&svc.ServiceContext{
		DB:     newGormTestDB(t),
		Config: &config.Config{},
	})
	if err != nil {
		t.Fatalf("NewAttachmentsController: %v", err)
	}
	return h
}

func mustNewAttachmentsControllerWithMock(t *testing.T) (*AttachmentsController, sqlmock.Sqlmock) {
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
	h, err := NewAttachmentsController(&svc.ServiceContext{
		DB:     db,
		Config: &config.Config{},
	})
	if err != nil {
		t.Fatalf("NewAttachmentsController: %v", err)
	}
	return h, mock
}

func mustNewAttachmentsControllerWithFakeDriver(t *testing.T) (*AttachmentsController, sqlmock.Sqlmock, *fakeAttachmentDriver) {
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
	fakeDrv := &fakeAttachmentDriver{}
	registry := driver.NewDriverRegistry()
	registry.Register("fake", func(json.RawMessage) (driver.DriverBackend, error) {
		return fakeDrv, nil
	})
	h, err := NewAttachmentsController(&svc.ServiceContext{
		DB:             db,
		Config:         &config.Config{},
		DriverStore:    driver.NewStore(db),
		DriverRegistry: registry,
	})
	if err != nil {
		t.Fatalf("NewAttachmentsController: %v", err)
	}
	return h, mock, fakeDrv
}

type fakeAttachmentDriver struct {
	savedKey    string
	deletedKeys []string
}

func (d *fakeAttachmentDriver) DriverName() string {
	return "fake"
}

func (d *fakeAttachmentDriver) Save(_ context.Context, req driver.PutRequest) (driver.StoredObject, error) {
	if req.Reader != nil {
		_, _ = io.Copy(io.Discard, req.Reader)
	}
	d.savedKey = req.ObjectKey
	return driver.StoredObject{
		LocalPath: req.ObjectKey,
		ETag:      "etag",
		Checksum:  "sha256:test",
	}, nil
}

func (d *fakeAttachmentDriver) PresignUpload(context.Context, driver.PresignRequest) (driver.PresignResult, error) {
	return driver.PresignResult{}, driver.ErrUnsupportedDriver
}

func (d *fakeAttachmentDriver) PresignDownload(context.Context, string, time.Duration) (driver.PresignResult, error) {
	return driver.PresignResult{}, driver.ErrUnsupportedDriver
}

func (d *fakeAttachmentDriver) LocalFilePath(localPath string) (string, error) {
	return localPath, nil
}

func (d *fakeAttachmentDriver) Delete(_ context.Context, objectKey string) error {
	d.deletedKeys = append(d.deletedKeys, objectKey)
	return nil
}

func (d *fakeAttachmentDriver) HealthCheck(context.Context) error {
	return nil
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
