package tags

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
)

func TestCreateTagSuccess(t *testing.T) {
	c, mock := newCrudTestController(t)
	reqBody := v1.CreateTagRequest{
		Name: "Go",
		Slug: "go",
	}
	body, _ := json.Marshal(reqBody)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "content"\."tags".*RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
	mock.ExpectCommit()

	ctx, rec := testCrudContext(http.MethodPost, "/api/v1/tags", bytes.NewReader(body))
	c.Create(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.CreateTagResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if resp.ID != 42 {
		t.Fatalf("id = %d, want 42", resp.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateTagInvalidBody(t *testing.T) {
	c, _ := newCrudTestController(t)
	body := []byte(`{"name":"Go"}`)

	ctx, rec := testCrudContext(http.MethodPost, "/api/v1/tags", bytes.NewReader(body))
	c.Create(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusBadRequest, "invalid request body")
}

func TestCreateTagDBUniqueViolation(t *testing.T) {
	c, mock := newCrudTestController(t)
	reqBody := v1.CreateTagRequest{
		Name: "Go",
		Slug: "go",
	}
	body, _ := json.Marshal(reqBody)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "content"\."tags".*RETURNING "id"`).
		WillReturnError(&pgconn.PgError{Code: "23505"})
	mock.ExpectRollback()

	ctx, rec := testCrudContext(http.MethodPost, "/api/v1/tags", bytes.NewReader(body))
	c.Create(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusConflict, "tag slug is already taken")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
