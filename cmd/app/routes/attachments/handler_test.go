package attachments

import (
	"context"
	"errors"
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
	pkgattachment "github.com/HappyLadySauce/Beehive-Blog/pkg/attachment"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/config"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

func TestAttachmentManagementRejectsMemberRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	issuer := newAttachmentTestIssuer(t)
	pair, err := issuer.IssuePair(42, "member")
	if err != nil {
		t.Fatalf("IssuePair: %v", err)
	}

	r := gin.New()
	g := r.Group("/api/v1/attachments")
	g.Use(middleware.AuthMiddleware(&svc.ServiceContext{Token: issuer}), middleware.RequireRole("admin"))
	g.POST("", func(ctx *gin.Context) {
		ctx.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/attachments", nil)
	req.Header.Set("Authorization", "Bearer "+pair.Access.Token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestAttachmentInitRejectsMissingDependencies(t *testing.T) {
	if err := Init(nil); err == nil || !strings.Contains(err.Error(), "service context is nil") {
		t.Fatalf("Init(nil) error = %v, want service context error", err)
	}

	err := Init(&svc.ServiceContext{Config: &config.Config{}})
	if err == nil || !strings.Contains(err.Error(), "attachment config is nil") {
		t.Fatalf("Init without attachment config error = %v, want attachment config error", err)
	}
}

func TestGetPublicContentRejectsPrivateAttachment(t *testing.T) {
	h, mock := newAttachmentDBTestController(t)
	expectAttachmentFirst(mock, attachmentRow{
		ID:           10,
		Purpose:      pkgattachment.PurposeContent,
		Filename:     "private.png",
		MimeType:     "image/png",
		Size:         5,
		StorageType:  options.AttachmentStorageLocal,
		LocalPath:    "content/private.png",
		AccessScope:  pkgattachment.AccessPrivate,
		UploadStatus: pkgattachment.UploadReady,
		Status:       pkgattachment.StatusActive,
	})
	expectCategoryBindings(mock)

	_, err := h.getPublicContent(context.Background(), 10, false)
	if !errors.Is(err, pkgattachment.ErrForbidden) {
		t.Fatalf("getPublicContent error = %v, want ErrForbidden", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestPatchRejectsPendingPublicAttachment(t *testing.T) {
	h, mock := newAttachmentDBTestController(t)
	public := pkgattachment.AccessPublic
	mock.ExpectBegin()
	expectAttachmentFirst(mock, attachmentRow{
		ID:           11,
		Purpose:      pkgattachment.PurposeContent,
		Filename:     "pending.png",
		MimeType:     "image/png",
		Size:         5,
		StorageType:  options.AttachmentStorageS3,
		ObjectKey:    "content/pending.png",
		AccessScope:  pkgattachment.AccessPrivate,
		UploadStatus: pkgattachment.UploadPending,
		Status:       pkgattachment.StatusActive,
	})
	mock.ExpectRollback()

	_, err := h.patch(context.Background(), pkgattachment.Actor{UID: 1, Role: pkgattachment.RoleAdmin}, 11, pkgattachment.PatchInput{
		AccessScope: &public,
	})
	if !errors.Is(err, pkgattachment.ErrInvalid) {
		t.Fatalf("patch error = %v, want ErrInvalid", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestReplaceCategoriesRejectsDuplicateIDs(t *testing.T) {
	h, mock := newAttachmentDBTestController(t)
	mock.ExpectBegin()
	expectAttachmentFirst(mock, attachmentRow{
		ID:           12,
		Purpose:      pkgattachment.PurposeContent,
		Filename:     "ready.png",
		MimeType:     "image/png",
		Size:         5,
		StorageType:  options.AttachmentStorageLocal,
		LocalPath:    "content/ready.png",
		AccessScope:  pkgattachment.AccessPrivate,
		UploadStatus: pkgattachment.UploadReady,
		Status:       pkgattachment.StatusActive,
	})
	mock.ExpectExec(`DELETE FROM "attachment"."attachment_categories"`).
		WithArgs(int64(12)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "attachment"."categories"`).
		WithArgs(int64(1), pkgattachment.CategoryStatusActive).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectRollback()

	err := h.replaceCategories(context.Background(), pkgattachment.Actor{UID: 1, Role: pkgattachment.RoleAdmin}, 12, []int64{1, 1})
	if !errors.Is(err, pkgattachment.ErrInvalid) {
		t.Fatalf("replaceCategories error = %v, want ErrInvalid", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func newAttachmentTestIssuer(t *testing.T) *jwt.Issuer {
	t.Helper()
	issuer, err := jwt.NewIssuer(&options.JWTOptions{
		Issuer:     "beehive-blog-test",
		Secret:     "0123456789abcdef0123456789abcdef",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	})
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}
	return issuer
}

func newAttachmentDBTestController(t *testing.T) (*AttachmentsController, sqlmock.Sqlmock) {
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
	opts := options.NewAttachmentOptions()
	opts.LocalRoot = t.TempDir()
	return NewController(&svc.ServiceContext{
		Config: &config.Config{Attachment: opts},
		DB:     db,
	}), mock
}

type attachmentRow struct {
	ID           int64
	Purpose      string
	Filename     string
	MimeType     string
	Size         int64
	StorageType  string
	ObjectKey    string
	LocalPath    string
	AccessScope  string
	UploadStatus string
	Status       string
}

func expectAttachmentFirst(mock sqlmock.Sqlmock, row attachmentRow) {
	now := time.Now()
	rows := sqlmock.NewRows(attachmentColumns()).AddRow(
		row.ID,
		nil,
		row.Purpose,
		row.Filename,
		nil,
		row.MimeType,
		row.Size,
		row.StorageType,
		nil,
		nullableString(row.ObjectKey),
		nullableString(row.LocalPath),
		nil,
		nil,
		row.AccessScope,
		row.UploadStatus,
		row.Status,
		now,
		now,
		nil,
	)
	mock.ExpectQuery(`SELECT \* FROM "attachment"."attachments"`).
		WithArgs(row.ID, 1).
		WillReturnRows(rows)
}

func expectCategoryBindings(mock sqlmock.Sqlmock, rows ...model.AttachmentCategoryBinding) {
	out := sqlmock.NewRows([]string{"attachment_id", "category_id", "created_at"})
	for _, row := range rows {
		out.AddRow(row.AttachmentID, row.CategoryID, row.CreatedAt)
	}
	mock.ExpectQuery(`SELECT \* FROM "attachment"."attachment_categories"`).
		WillReturnRows(out)
}

func attachmentColumns() []string {
	return []string{
		"id",
		"owner_user_id",
		"purpose",
		"filename",
		"original_name",
		"mime_type",
		"size",
		"storage_type",
		"bucket",
		"object_key",
		"local_path",
		"etag",
		"checksum",
		"access_scope",
		"upload_status",
		"status",
		"created_at",
		"updated_at",
		"deleted_at",
	}
}

func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
