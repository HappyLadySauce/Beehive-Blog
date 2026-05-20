package tags

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
)

func TestListTagsPublicFiltersArchived(t *testing.T) {
	c, mock := newCrudTestController(t)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "content"."tags" WHERE`).
		WithArgs("active").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM "content"."tags" WHERE`).
		WithArgs("active", 10).
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

func TestListTagsAdminViaBearerCanFilterArchived(t *testing.T) {
	c, mock, issuer := newCrudTestControllerWithToken(t)
	pair, err := issuer.IssuePair(1, "admin")
	if err != nil {
		t.Fatalf("IssuePair: %v", err)
	}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "content"."tags" WHERE`).
		WithArgs("archived").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT \* FROM "content"."tags" WHERE`).
		WithArgs("archived", 10).
		WillReturnRows(sqlmock.NewRows(tagColumns()))

	ctx, rec := testCrudContext(http.MethodGet, "/api/v1/tags?status=archived", nil)
	ctx.Request.Header.Set("Authorization", "Bearer "+pair.Access.Token)
	runOptionalAuth(t, c, ctx, c.List)
	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
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
		WithArgs("archived", 10).
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
