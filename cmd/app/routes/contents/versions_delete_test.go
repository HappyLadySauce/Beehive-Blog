package contents

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestDeleteVersionSuccess(t *testing.T) {
	c, mock := newCrudTestController(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "T", "s", nil, nil, nil, 10, "draft", "public", "allowed", nil, 0, 0, []byte("{}"), int64(0), now, now, nil))
	mock.ExpectQuery(`SELECT \* FROM "content"."content_versions" WHERE`).
		WillReturnRows(sqlmock.NewRows(versionColumns()).
			AddRow(10, 1, 2, "manual", "v2", "T", nil, nil, nil, int64(1), now))
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "content"."content_versions" WHERE`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx, rec := testCrudContextWithID(http.MethodDelete, "/api/v1/contents/1/versions/2", nil, "1")
	ctx.Params = append(ctx.Params, gin.Param{Key: "versionNumber", Value: "2"})

	c.DeleteVersion(ctx)
	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteVersionContentNotFound(t *testing.T) {
	c, mock := newCrudTestController(t)

	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WillReturnRows(sqlmock.NewRows(contentColumns()))

	ctx, rec := testCrudContextWithID(http.MethodDelete, "/api/v1/contents/999/versions/2", nil, "999")
	ctx.Params = append(ctx.Params, gin.Param{Key: "versionNumber", Value: "2"})

	c.DeleteVersion(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusNotFound, "content not found")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteVersionNotFound(t *testing.T) {
	c, mock := newCrudTestController(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "T", "s", nil, nil, nil, 10, "draft", "public", "allowed", nil, 0, 0, []byte("{}"), int64(0), now, now, nil))
	mock.ExpectQuery(`SELECT \* FROM "content"."content_versions" WHERE`).
		WillReturnRows(sqlmock.NewRows(versionColumns()))

	ctx, rec := testCrudContextWithID(http.MethodDelete, "/api/v1/contents/1/versions/99", nil, "1")
	ctx.Params = append(ctx.Params, gin.Param{Key: "versionNumber", Value: "99"})

	c.DeleteVersion(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusNotFound, "version not found")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteVersionRejectsAuto(t *testing.T) {
	c, mock := newCrudTestController(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "T", "s", nil, nil, nil, 10, "draft", "public", "allowed", nil, 0, 0, []byte("{}"), int64(0), now, now, nil))
	mock.ExpectQuery(`SELECT \* FROM "content"."content_versions" WHERE`).
		WillReturnRows(sqlmock.NewRows(versionColumns()).
			AddRow(10, 1, 1, "auto", "自动保存", "T", nil, nil, nil, int64(1), now))

	ctx, rec := testCrudContextWithID(http.MethodDelete, "/api/v1/contents/1/versions/1", nil, "1")
	ctx.Params = append(ctx.Params, gin.Param{Key: "versionNumber", Value: "1"})

	c.DeleteVersion(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusConflict, "auto snapshots cannot be deleted")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteVersionInvalidVersionNumber(t *testing.T) {
	c, _ := newCrudTestController(t)

	ctx, rec := testCrudContextWithID(http.MethodDelete, "/api/v1/contents/1/versions/abc", nil, "1")
	ctx.Params = append(ctx.Params, gin.Param{Key: "versionNumber", Value: "abc"})

	c.DeleteVersion(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusBadRequest, "versionNumber must be a positive integer")
}
