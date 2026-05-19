package users

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDeleteUserNotFound(t *testing.T) {
	c, mock := newCrudTestController(t)
	mock.ExpectQuery(`SELECT .* FROM "identity"\."users".*"users"\."id" = \$1.*LIMIT \$2`).
		WithArgs(99, 1).
		WillReturnRows(sqlmock.NewRows(userColumns()))

	ctx, rec := testCrudContextWithID(http.MethodDelete, "/api/v1/users/99", nil, "99")
	c.Delete(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusNotFound, "user not found")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteUserSuccess(t *testing.T) {
	c, mock := newCrudTestController(t)
	mock.ExpectQuery(`SELECT .* FROM "identity"\."users".*"users"\."id" = \$1.*LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(1, "alice", nil, nil, nil, nil, "member", "active", nil, time.Now(), time.Now(), nil))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "identity"\."users"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE "identity"\."user_credentials"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE "identity"\."user_identities"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE "identity"\."user_sessions"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx, rec := testCrudContextWithID(http.MethodDelete, "/api/v1/users/1", nil, "1")
	c.Delete(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
