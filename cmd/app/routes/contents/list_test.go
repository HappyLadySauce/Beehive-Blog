package contents

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListContentsSlugFilter(t *testing.T) {
	c, mock := newCrudTestController(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "content"."contents" WHERE`).
		WithArgs("published", "public", "article", "my-slug").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WithArgs("published", "public", "article", "my-slug", 10).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "T", "my-slug", nil, nil, nil, 10, "published", "public", "allowed", &now, 0, 0, []byte("{}"), int64(0), now, now, nil))
	mock.ExpectQuery(`SELECT id, username FROM "identity"."users" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(10, "author"))
	mock.ExpectQuery(`SELECT \* FROM "content"."content_tags" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"content_id", "tag_id"}))

	ctx, rec := testCrudContext(http.MethodGet, "/api/v1/contents?slug=my-slug&type=article", nil)
	c.List(ctx)
	env := decodeCrudEnvelope(t, rec)
	if rec.Code != http.StatusOK || env.Code != 200 {
		t.Fatalf("HTTP=%d code=%d", rec.Code, env.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListContentsBatchLoadAuthorFailure(t *testing.T) {
	c, mock := newCrudTestController(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "content"."contents" WHERE`).
		WithArgs("published", "public").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WithArgs("published", "public", 10).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "T", "s", nil, nil, nil, 10, "published", "public", "allowed", &now, 0, 0, []byte("{}"), int64(0), now, now, nil))
	mock.ExpectQuery(`SELECT id, username FROM "identity"."users" WHERE`).
		WillReturnError(sqlmock.ErrCancelled)

	ctx, rec := testCrudContext(http.MethodGet, "/api/v1/contents", nil)
	c.List(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusInternalServerError, "failed to load author usernames")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
