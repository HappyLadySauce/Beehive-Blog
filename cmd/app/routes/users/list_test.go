package users_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
)

func TestListUsersEmpty(t *testing.T) {
	c, mock := newCrudTestController(t)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "identity"\."users"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT .* FROM "identity"\."users".*ORDER BY id DESC LIMIT \$1`).
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows(userColumns()))

	ctx, rec := testCrudContext(http.MethodGet, "/api/v1/users", nil)
	c.List(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.ListUsersResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if len(resp.Items) != 0 {
		t.Fatalf("items = %d, want 0", len(resp.Items))
	}
	if resp.Total != 0 {
		t.Fatalf("total = %d, want 0", resp.Total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListUsersWithResults(t *testing.T) {
	c, mock := newCrudTestController(t)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "identity"\."users"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(`SELECT .* FROM "identity"\."users".*ORDER BY id DESC LIMIT \$1`).
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(2, "bob", "bob@test.com", nil, nil, nil, "member", "active", nil, time.Now(), time.Now(), nil).
			AddRow(1, "alice", nil, nil, nil, nil, "admin", "active", nil, time.Now(), time.Now(), nil))

	ctx, rec := testCrudContext(http.MethodGet, "/api/v1/users", nil)
	c.List(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.ListUsersResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(resp.Items))
	}
	if resp.Total != 2 {
		t.Fatalf("total = %d, want 2", resp.Total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
