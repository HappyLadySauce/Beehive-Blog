package users

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
)

func TestUpdateUserNotFound(t *testing.T) {
	c, mock := newCrudTestController(t)
	mock.ExpectQuery(`SELECT .* FROM "identity"\."users".*"users"\."id" = \$1.*LIMIT \$2`).
		WithArgs(99, 1).
		WillReturnRows(sqlmock.NewRows(userColumns()))

	reqBody, _ := json.Marshal(map[string]interface{}{"nickname": "newname"})
	ctx, rec := testCrudContextWithID(http.MethodPatch, "/api/v1/users/99", bytes.NewReader(reqBody), "99")
	c.Update(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusNotFound, "user not found")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateUserSuccess(t *testing.T) {
	c, mock := newCrudTestController(t)
	mock.ExpectQuery(`SELECT .* FROM "identity"\."users".*"users"\."id" = \$1.*LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(1, "alice", "alice@test.com", nil, nil, nil, "member", "active", nil, time.Now(), time.Now(), nil))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "identity"\."users" SET .*WHERE`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT .* FROM "identity"\."users"`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(1, "alice", "alice@test.com", nil, nil, nil, "admin", "active", nil, time.Now(), time.Now(), nil))

	reqBody, _ := json.Marshal(map[string]interface{}{"role": "admin"})
	ctx, rec := testCrudContextWithID(http.MethodPatch, "/api/v1/users/1", bytes.NewReader(reqBody), "1")
	c.Update(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.UserDetailResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if resp.Role != "admin" {
		t.Fatalf("role = %q, want admin", resp.Role)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateUserClearEmail(t *testing.T) {
	c, mock := newCrudTestController(t)
	mock.ExpectQuery(`SELECT .* FROM "identity"\."users".*"users"\."id" = \$1.*LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(1, "alice", "old@test.com", nil, nil, nil, "member", "active", nil, time.Now(), time.Now(), nil))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "identity"\."users" SET .*WHERE`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT .* FROM "identity"\."users"`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(1, "alice", nil, nil, nil, nil, "member", "active", nil, time.Now(), time.Now(), nil))

	reqBody, _ := json.Marshal(map[string]interface{}{"email": ""})
	ctx, rec := testCrudContextWithID(http.MethodPatch, "/api/v1/users/1", bytes.NewReader(reqBody), "1")
	c.Update(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.UserDetailResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if resp.Email != nil {
		t.Fatalf("email = %v, want nil", resp.Email)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
