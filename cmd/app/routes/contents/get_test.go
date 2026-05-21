package contents

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
)

func TestGetContentNotFound(t *testing.T) {
	c, mock := newCrudTestController(t)

	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WithArgs(99, "published", "public", 1).
		WillReturnRows(sqlmock.NewRows(contentColumns()))

	ctx, rec := testCrudContextWithID(http.MethodGet, "/api/v1/contents/99", nil, "99")
	c.Get(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusNotFound, "content not found")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetContentDraftHiddenFromPublic(t *testing.T) {
	c, mock := newCrudTestController(t)

	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WithArgs(1, "published", "public", 1).
		WillReturnRows(sqlmock.NewRows(contentColumns()))

	ctx, rec := testCrudContextWithID(http.MethodGet, "/api/v1/contents/1", nil, "1")
	c.Get(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusNotFound, "content not found")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetContentPublicReturnsPublicFields(t *testing.T) {
	c, mock := newCrudTestController(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WithArgs(1, "published", "public", 1).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "Test", "test-slug", nil, nil, nil, 10, "published", "public", "allowed", &now, 100, 2, json.RawMessage("{}"), int64(50), now, now, nil))

	// GORM Select("username").First(&user, id) generates: SELECT "username" FROM "identity"."users" WHERE ...
	mock.ExpectQuery(`SELECT "username" FROM "identity"."users" WHERE`).
		WithArgs(10, 1).
		WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("author1"))

	// Atomic view count increment (raw SQL Exec).
	mock.ExpectExec(`UPDATE content.contents SET view_count`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Content tags: SELECT * FROM "content"."content_tags" WHERE content_id = ?
	mock.ExpectQuery(`SELECT \* FROM "content"."content_tags" WHERE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"content_id", "tag_id", "created_at"}))

	// Content categories: SELECT * FROM "content"."content_categories" WHERE content_id = ?
	mock.ExpectQuery(`SELECT \* FROM "content"."content_categories" WHERE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"content_id", "category_id", "created_at"}))

	ctx, rec := testCrudContextWithID(http.MethodGet, "/api/v1/contents/1", nil, "1")
	c.Get(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}

	var resp v1.PublicContentDetailResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if resp.ID != 1 {
		t.Fatalf("id = %d, want 1", resp.ID)
	}
	if resp.Title != "Test" {
		t.Fatalf("title = %q, want Test", resp.Title)
	}
	if resp.AuthorUsername != "author1" {
		t.Fatalf("author_username = %q, want author1", resp.AuthorUsername)
	}
	if strings.Contains(string(env.Data), `"metadata"`) {
		t.Fatalf("public response must not include metadata, got %s", env.Data)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetContentAdminViaBearerToken(t *testing.T) {
	c, mock, issuer := newCrudTestControllerWithToken(t)
	now := time.Now()
	pair, err := issuer.IssuePair(10, "admin")
	if err != nil {
		t.Fatalf("IssuePair: %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "Draft", "draft-slug", nil, nil, nil, 10, "draft", "private", "denied", nil, 0, 0, json.RawMessage(`{"secret":true}`), int64(0), now, now, nil))

	mock.ExpectQuery(`SELECT "username" FROM "identity"."users" WHERE`).
		WithArgs(10, 1).
		WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("author1"))

	mock.ExpectQuery(`SELECT \* FROM "content"."content_tags" WHERE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"content_id", "tag_id", "created_at"}))

	mock.ExpectQuery(`SELECT \* FROM "content"."content_categories" WHERE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"content_id", "category_id", "created_at"}))

	ctx, rec := testCrudContextWithID(http.MethodGet, "/api/v1/contents/1", nil, "1")
	ctx.Request.Header.Set("Authorization", "Bearer "+pair.Access.Token)
	runOptionalAuth(t, c, ctx, c.Get)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.ContentDetailResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if resp.Status != "draft" {
		t.Fatalf("status = %q, want draft", resp.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetContentAdminReturnsFullFields(t *testing.T) {
	c, mock := newCrudTestController(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "Test", "test-slug", nil, nil, nil, 10, "draft", "private", "denied", nil, 0, 0, json.RawMessage("{}"), int64(0), now, now, nil))

	mock.ExpectQuery(`SELECT "username" FROM "identity"."users" WHERE`).
		WithArgs(10, 1).
		WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("author1"))

	mock.ExpectQuery(`SELECT \* FROM "content"."content_tags" WHERE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"content_id", "tag_id", "created_at"}))

	mock.ExpectQuery(`SELECT \* FROM "content"."content_categories" WHERE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"content_id", "category_id", "created_at"}))

	ctx, rec := testCrudContextWithID(http.MethodGet, "/api/v1/contents/1", nil, "1")
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 10, Role: "admin"})

	c.Get(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}

	var resp v1.ContentDetailResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if resp.ID != 1 {
		t.Fatalf("id = %d, want 1", resp.ID)
	}
	if resp.Status != "draft" {
		t.Fatalf("status = %q, want draft", resp.Status)
	}
	if resp.Visibility != "private" {
		t.Fatalf("visibility = %q, want private", resp.Visibility)
	}
	if resp.AIAccess != "denied" {
		t.Fatalf("ai_access = %q, want denied", resp.AIAccess)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetContentPublicLoadsLinkedTags(t *testing.T) {
	c, mock := newCrudTestController(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WithArgs(1, "published", "public", 1).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "Test", "test-slug", nil, nil, nil, 10, "published", "public", "allowed", &now, 100, 2, json.RawMessage("{}"), int64(50), now, now, nil))

	mock.ExpectQuery(`SELECT "username" FROM "identity"."users" WHERE`).
		WithArgs(10, 1).
		WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("author1"))

	mock.ExpectExec(`UPDATE content.contents SET view_count`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(`SELECT \* FROM "content"."content_tags" WHERE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"content_id", "tag_id", "created_at"}).
			AddRow(1, 1, now).
			AddRow(1, 2, now))

	mock.ExpectQuery(`SELECT \* FROM "content"."tags" WHERE`).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows(tagColumns()).
			AddRow(1, "Active", "active", nil, nil, now, now, nil))

	mock.ExpectQuery(`SELECT \* FROM "content"."content_categories" WHERE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"content_id", "category_id", "created_at"}).
			AddRow(1, 10, now))

	mock.ExpectQuery(`SELECT \* FROM "content"."categories" WHERE`).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows(categoryColumns()).
			AddRow(10, "Tech", "tech", nil, nil, 0, now, now, nil))

	ctx, rec := testCrudContextWithID(http.MethodGet, "/api/v1/contents/1", nil, "1")
	c.Get(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.PublicContentDetailResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if len(resp.Tags) != 1 {
		t.Fatalf("tags len = %d, want 1 (missing tag rows omitted)", len(resp.Tags))
	}
	if resp.Tags[0].Slug != "active" {
		t.Fatalf("tag slug = %q, want active", resp.Tags[0].Slug)
	}
	if len(resp.Categories) != 1 {
		t.Fatalf("categories len = %d, want 1", len(resp.Categories))
	}
	if resp.Categories[0].Slug != "tech" {
		t.Fatalf("category slug = %q, want tech", resp.Categories[0].Slug)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetContentTagLoadFailure(t *testing.T) {
	c, mock := newCrudTestController(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WithArgs(1, "published", "public", 1).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "Test", "test-slug", nil, nil, nil, 10, "published", "public", "allowed", &now, 100, 2, json.RawMessage("{}"), int64(50), now, now, nil))

	mock.ExpectQuery(`SELECT "username" FROM "identity"."users" WHERE`).
		WithArgs(10, 1).
		WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("author1"))

	mock.ExpectExec(`UPDATE content.contents SET view_count`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(`SELECT \* FROM "content"."content_tags" WHERE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"content_id", "tag_id", "created_at"}).
			AddRow(1, 1, now))

	mock.ExpectQuery(`SELECT \* FROM "content"."tags" WHERE`).
		WillReturnError(sqlmock.ErrCancelled)

	ctx, rec := testCrudContextWithID(http.MethodGet, "/api/v1/contents/1", nil, "1")
	c.Get(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusInternalServerError, "failed to load content tags")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetContentViewCountFailureStillOK(t *testing.T) {
	c, mock := newCrudTestController(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WithArgs(1, "published", "public", 1).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "Test", "test-slug", nil, nil, nil, 10, "published", "public", "allowed", &now, 100, 2, json.RawMessage("{}"), int64(50), now, now, nil))

	mock.ExpectQuery(`SELECT "username" FROM "identity"."users" WHERE`).
		WithArgs(10, 1).
		WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("author1"))

	mock.ExpectExec(`UPDATE content.contents SET view_count`).
		WithArgs(int64(1)).
		WillReturnError(sqlmock.ErrCancelled)

	mock.ExpectQuery(`SELECT \* FROM "content"."content_tags" WHERE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"content_id", "tag_id", "created_at"}))

	mock.ExpectQuery(`SELECT \* FROM "content"."content_categories" WHERE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"content_id", "category_id", "created_at"}))

	ctx, rec := testCrudContextWithID(http.MethodGet, "/api/v1/contents/1", nil, "1")
	c.Get(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200 when view_count update fails", rec.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func tagColumns() []string {
	return []string{"id", "name", "slug", "description", "color", "created_at", "updated_at", "deleted_at"}
}

func categoryColumns() []string {
	return []string{"id", "name", "slug", "description", "parent_id", "sort_order", "created_at", "updated_at", "deleted_at"}
}
