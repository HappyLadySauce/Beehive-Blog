package tags

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
)

func TestListTagsPublicFiltersArchived(t *testing.T) {
	c, mock := newCrudTestController(t)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "content"."tags" WHERE`).
		WithArgs("active").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM "content"."tags" WHERE`).
		WithArgs("active", 20).
		WillReturnRows(sqlmock.NewRows(tagColumns()).
			AddRow(1, "Go", "go", nil, nil, "active", now, now, nil))

	mock.ExpectQuery(`SELECT tag_id, COUNT\(\*\) as count FROM "content"."content_tags" WHERE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"tag_id", "count"}))

	ctx, rec := testCrudContext(http.MethodGet, "/api/v1/tags?status=archived", nil)
	c.List(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.ListTagsResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if resp.Total != 2 {
		t.Fatalf("total = %d, want 2", resp.Total)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("items len = %d, want 1", len(resp.Items))
	}
	if resp.Items[0].Status != "" {
		t.Fatalf("status = %q, want empty (public omits status)", resp.Items[0].Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListTagsAdminCanFilterByStatus(t *testing.T) {
	c, mock := newCrudTestController(t)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "content"."tags" WHERE`).
		WithArgs("archived").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM "content"."tags" WHERE`).
		WithArgs("archived", 20).
		WillReturnRows(sqlmock.NewRows(tagColumns()).
			AddRow(2, "OldTag", "old-tag", nil, nil, "archived", now, now, nil))

	mock.ExpectQuery(`SELECT tag_id, COUNT\(\*\) as count FROM "content"."content_tags" WHERE`).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"tag_id", "count"}))

	ctx, rec := testCrudContext(http.MethodGet, "/api/v1/tags?status=archived", nil)
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 10, Role: "admin"})

	c.List(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.ListTagsResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if len(resp.Items) > 0 && resp.Items[0].Status != "archived" {
		t.Fatalf("status = %q, want archived", resp.Items[0].Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetTagPublicOmitsStatus(t *testing.T) {
	c, mock := newCrudTestController(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM "content"."tags" WHERE`).
		WithArgs(int64(1), 1).
		WillReturnRows(sqlmock.NewRows(tagColumns()).
			AddRow(1, "Go", "go", nil, nil, "active", now, now, nil))

	mock.ExpectQuery(`SELECT count\(\*\) FROM "content"."content_tags" WHERE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	ctx, rec := testCrudContextWithID(http.MethodGet, "/api/v1/tags/1", nil, "1")
	c.Get(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.TagDetailResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if resp.ID != 1 {
		t.Fatalf("id = %d, want 1", resp.ID)
	}
	if resp.Name != "Go" {
		t.Fatalf("name = %q, want Go", resp.Name)
	}
	if resp.Status != "" {
		t.Fatalf("status = %q, want empty (public omits status)", resp.Status)
	}
	if resp.ContentCount != 5 {
		t.Fatalf("content_count = %d, want 5", resp.ContentCount)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestMapTagCrudUniqueViolation(t *testing.T) {
	// Direct unit test for the error mapper since sqlmock+GORM doesn't
	// reliably propagate pgconn.PgError through the driver layer.
	// 直接测试错误映射函数，因为 sqlmock+GORM 无法可靠地传播 pgconn.PgError。
	pgErr := &pgconn.PgError{Code: "23505"}
	err := mapTagCrudUniqueViolation(pgErr, "my-slug")

	appErr, ok := err.(*common.AppError)
	if !ok {
		t.Fatalf("expected *common.AppError, got %T", err)
	}
	if appErr.HTTPStatus != http.StatusConflict {
		t.Fatalf("HTTP status = %d, want 409", appErr.HTTPStatus)
	}
	if appErr.Message != "my-slug slug is already taken" {
		t.Fatalf("message = %q, want 'my-slug slug is already taken'", appErr.Message)
	}
}

func TestMapTagCrudUniqueViolationNonPgError(t *testing.T) {
	err := mapTagCrudUniqueViolation(sqlmock.ErrCancelled, "test")

	appErr, ok := err.(*common.AppError)
	if !ok {
		t.Fatalf("expected *common.AppError, got %T", err)
	}
	if appErr.HTTPStatus != http.StatusInternalServerError {
		t.Fatalf("HTTP status = %d, want 500", appErr.HTTPStatus)
	}
}

// TestUpdateTagSlugErrorVerifiedDirectly is covered by TestMapTagCrudUniqueViolation
// and TestMapTagCrudUniqueViolationNonPgError which test the error mapper directly.
// Full-flow mock test is skipped because sqlmock+GORM does not reliably propagate
// pgconn.PgError through the driver layer for Exec-based operations.

func strPtr(s string) *string { return &s }
