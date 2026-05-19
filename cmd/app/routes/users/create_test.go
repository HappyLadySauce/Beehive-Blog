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

func TestCreateUserSuccess(t *testing.T) {
	c, mock := newCrudTestController(t)
	mock.ExpectQuery(`SELECT .* FROM "identity"\."users".*username = \$1 AND id != \$2.*LIMIT \$3`).
		WithArgs("charlie", int64(0), 1).
		WillReturnRows(sqlmock.NewRows(userColumns()))
	mock.ExpectQuery(`SELECT .* FROM "identity"\."users".*email = \$1 AND id != \$2.*LIMIT \$3`).
		WithArgs("charlie@test.com", int64(0), 1).
		WillReturnRows(sqlmock.NewRows(userColumns()))
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "identity"\."users".*RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
	mock.ExpectQuery(`INSERT INTO "identity"\."user_credentials".*RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
	mock.ExpectCommit()

	reqBody, _ := json.Marshal(map[string]interface{}{
		"username": "charlie",
		"password": "password123",
		"email":    "charlie@test.com",
	})
	ctx, rec := testCrudContext(http.MethodPost, "/api/v1/users", bytes.NewReader(reqBody))
	c.Create(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.CreateUserResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if resp.ID != 3 {
		t.Fatalf("id = %d, want 3", resp.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateUserNoPassword(t *testing.T) {
	c, mock := newCrudTestController(t)
	mock.ExpectQuery(`SELECT .* FROM "identity"\."users".*username = \$1 AND id != \$2.*LIMIT \$3`).
		WithArgs("oauthuser", int64(0), 1).
		WillReturnRows(sqlmock.NewRows(userColumns()))
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "identity"\."users".*RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(4))
	mock.ExpectCommit()

	reqBody, _ := json.Marshal(map[string]interface{}{
		"username": "oauthuser",
	})
	ctx, rec := testCrudContext(http.MethodPost, "/api/v1/users", bytes.NewReader(reqBody))
	c.Create(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.CreateUserResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if resp.ID != 4 {
		t.Fatalf("id = %d, want 4", resp.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateUserUsernameConflict(t *testing.T) {
	c, mock := newCrudTestController(t)
	mock.ExpectQuery(`SELECT .* FROM "identity"\."users".*username = \$1 AND id != \$2.*LIMIT \$3`).
		WithArgs("bob", int64(0), 1).
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(2, "bob", nil, nil, nil, nil, "member", "active", nil, time.Now(), time.Now(), nil))

	reqBody, _ := json.Marshal(map[string]interface{}{
		"username": "bob",
		"password": "password123",
	})
	ctx, rec := testCrudContext(http.MethodPost, "/api/v1/users", bytes.NewReader(reqBody))
	c.Create(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusConflict, "username is already taken")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateUserEmailConflict(t *testing.T) {
	c, mock := newCrudTestController(t)
	mock.ExpectQuery(`SELECT .* FROM "identity"\."users".*username = \$1 AND id != \$2.*LIMIT \$3`).
		WithArgs("newuser", int64(0), 1).
		WillReturnRows(sqlmock.NewRows(userColumns()))
	mock.ExpectQuery(`SELECT .* FROM "identity"\."users".*email = \$1 AND id != \$2.*LIMIT \$3`).
		WithArgs("taken@test.com", int64(0), 1).
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(5, "other", "taken@test.com", nil, nil, nil, "member", "active", nil, time.Now(), time.Now(), nil))

	reqBody, _ := json.Marshal(map[string]interface{}{
		"username": "newuser",
		"password": "password123",
		"email":    "taken@test.com",
	})
	ctx, rec := testCrudContext(http.MethodPost, "/api/v1/users", bytes.NewReader(reqBody))
	c.Create(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusConflict, "email is already registered")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
