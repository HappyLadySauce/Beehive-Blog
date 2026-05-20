package tags

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
)

func TestGetTagPublicOmitsStatus(t *testing.T) {
	c, mock := newCrudTestController(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM "content"."tags" WHERE`).
		WithArgs(int64(1), "active", 1).
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

func TestGetTagArchivedHiddenFromPublic(t *testing.T) {
	c, mock := newCrudTestController(t)

	mock.ExpectQuery(`SELECT \* FROM "content"."tags" WHERE`).
		WithArgs(int64(2), "active", 1).
		WillReturnRows(sqlmock.NewRows(tagColumns()))

	ctx, rec := testCrudContextWithID(http.MethodGet, "/api/v1/tags/2", nil, "2")
	c.Get(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusNotFound, "tag not found")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
