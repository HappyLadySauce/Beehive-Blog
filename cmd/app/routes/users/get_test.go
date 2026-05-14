package users_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
)

func TestGetUserNotFound(t *testing.T) {
	c, mock := newCrudTestController(t)
	mock.ExpectQuery(`SELECT .* FROM "identity"\."users".*"users"\."id" = \$1.*LIMIT \$2`).
		WithArgs(99, 1).
		WillReturnRows(sqlmock.NewRows(userColumns()))

	ctx, rec := testCrudContextWithID(http.MethodGet, "/api/v1/users/99", nil, "99")
	c.Get(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusNotFound, "user not found")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetUserSuccess(t *testing.T) {
	c, mock := newCrudTestController(t)
	mock.ExpectQuery(`SELECT .* FROM "identity"\."users".*"users"\."id" = \$1.*LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(1, "alice", "alice@test.com", nil, nil, nil, "admin", "active", nil, time.Now(), time.Now(), nil))

	ctx, rec := testCrudContextWithID(http.MethodGet, "/api/v1/users/1", nil, "1")
	c.Get(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.UserDetailResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if resp.ID != 1 {
		t.Fatalf("id = %d, want 1", resp.ID)
	}
	if resp.Username != "alice" {
		t.Fatalf("username = %q, want alice", resp.Username)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
