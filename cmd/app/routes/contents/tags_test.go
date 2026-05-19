package contents

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
)

func TestSetContentTagsRejectsDuplicateIDs(t *testing.T) {
	c, _ := newCrudTestController(t)
	body, _ := json.Marshal(v1.SetContentTagsRequest{TagIDs: []int64{1, 1}})

	ctx, rec := testCrudContextWithID(http.MethodPut, "/api/v1/contents/1/tags", bytes.NewReader(body), "1")
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 10, Role: "admin"})

	c.SetTags(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusBadRequest, "duplicate tag IDs are not allowed")
}

func TestSetContentTagsAllowsEmptyList(t *testing.T) {
	c, mock := newCrudTestController(t)
	body, _ := json.Marshal(v1.SetContentTagsRequest{TagIDs: []int64{}})

	mock.ExpectQuery(`SELECT count\(\*\) FROM "content"."contents" WHERE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "content"."content_tags" WHERE`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	ctx, rec := testCrudContextWithID(http.MethodPut, "/api/v1/contents/1/tags", bytes.NewReader(body), "1")
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 10, Role: "admin"})

	c.SetTags(ctx)
	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
