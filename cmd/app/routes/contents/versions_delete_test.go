package contents

import (
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestDeleteVersionSuccess(t *testing.T) {
	c, mock := newCrudTestController(t)

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

func TestDeleteVersionNotFound(t *testing.T) {
	c, mock := newCrudTestController(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "content"."content_versions" WHERE`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	ctx, rec := testCrudContextWithID(http.MethodDelete, "/api/v1/contents/1/versions/99", nil, "1")
	ctx.Params = append(ctx.Params, gin.Param{Key: "versionNumber", Value: "99"})

	c.DeleteVersion(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusNotFound, "version not found")
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
