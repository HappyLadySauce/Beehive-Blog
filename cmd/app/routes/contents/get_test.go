package contents

import (
	"encoding/json"
	"net/http"
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
