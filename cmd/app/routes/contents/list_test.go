package contents

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListContentsBatchLoadAuthorFailure(t *testing.T) {
	c, mock := newCrudTestController(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "content"."contents" WHERE`).
		WithArgs("published", "public").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WithArgs("published", "public", 20).
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
