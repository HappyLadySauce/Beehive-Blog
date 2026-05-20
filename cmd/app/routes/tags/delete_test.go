package tags

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDeleteTagNotFound(t *testing.T) {
	c, mock := newCrudTestController(t)

	mock.ExpectQuery(`SELECT \* FROM "content"."tags" WHERE`).
		WithArgs(int64(99), 1).
		WillReturnRows(sqlmock.NewRows(tagColumns()))

	ctx, rec := testCrudContextWithID(http.MethodDelete, "/api/v1/tags/99", nil, "99")
	c.Delete(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusNotFound, "tag not found")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteTagReferencedConflict(t *testing.T) {
	c, mock := newCrudTestController(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM "content"."tags" WHERE`).
		WithArgs(int64(1), 1).
		WillReturnRows(sqlmock.NewRows(tagColumns()).
			AddRow(1, "Go", "go", nil, nil, "active", now, now, nil))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "content"."content_tags" WHERE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	ctx, rec := testCrudContextWithID(http.MethodDelete, "/api/v1/tags/1", nil, "1")
	c.Delete(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusConflict, "tag is referenced by 3 content item(s); remove references first")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteTagSuccess(t *testing.T) {
	c, mock := newCrudTestController(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM "content"."tags" WHERE`).
		WithArgs(int64(1), 1).
		WillReturnRows(sqlmock.NewRows(tagColumns()).
			AddRow(1, "Go", "go", nil, nil, "active", now, now, nil))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "content"."content_tags" WHERE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "content"\."tags" SET "deleted_at"=\$1 WHERE "tags"\."id" = \$2 AND "tags"\."deleted_at" IS NULL`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx, rec := testCrudContextWithID(http.MethodDelete, "/api/v1/tags/1", nil, "1")
	c.Delete(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteTagInvalidID(t *testing.T) {
	c, _ := newCrudTestController(t)

	ctx, rec := testCrudContextWithID(http.MethodDelete, "/api/v1/tags/abc", nil, "abc")
	c.Delete(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusBadRequest, "invalid tag id")
}
