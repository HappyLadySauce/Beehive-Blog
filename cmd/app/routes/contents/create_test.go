package contents

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
)

func TestCreateContentSlugConflict(t *testing.T) {
	c, mock := newCrudTestController(t)
	reqBody := v1.CreateContentRequest{
		Type:  "article",
		Title: "Test Article",
		Slug:  "duplicate-slug",
	}
	body, _ := json.Marshal(reqBody)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT count\(\*\) FROM "content"."contents" WHERE`).
		WithArgs("article", "duplicate-slug").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectRollback()

	ctx, rec := testCrudContext(http.MethodPost, "/api/v1/contents", bytes.NewReader(body))
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 10, Role: "admin"})

	c.Create(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusConflict, "slug \"duplicate-slug\" is already taken for type \"article\"")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateContentSuccess(t *testing.T) {
	c, mock := newCrudTestController(t)
	reqBody := v1.CreateContentRequest{
		Type:  "article",
		Title: "New Article",
		Slug:  "new-article",
	}
	body, _ := json.Marshal(reqBody)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT count\(\*\) FROM "content"."contents" WHERE`).
		WithArgs("article", "new-article").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	// GORM INSERT with RETURNING "metadata","id" (json.RawMessage gets special GORM scanner treatment).
	mock.ExpectQuery(`INSERT INTO "content"."contents"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(100))
	mock.ExpectCommit()

	ctx, rec := testCrudContext(http.MethodPost, "/api/v1/contents", bytes.NewReader(body))
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 10, Role: "admin"})

	c.Create(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.CreateContentResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if resp.ID != 100 {
		t.Fatalf("id = %d, want 100", resp.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateContentUnauthorized(t *testing.T) {
	c, _ := newCrudTestController(t)
	reqBody := v1.CreateContentRequest{
		Type:  "article",
		Title: "Test",
		Slug:  "test",
	}
	body, _ := json.Marshal(reqBody)

	ctx, rec := testCrudContext(http.MethodPost, "/api/v1/contents", bytes.NewReader(body))
	c.Create(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusUnauthorized, "authentication required")
}

func TestCreateContentDBUniqueViolation(t *testing.T) {
	c, mock := newCrudTestController(t)
	reqBody := v1.CreateContentRequest{
		Type:  "article",
		Title: "Test",
		Slug:  "duplicate",
	}
	body, _ := json.Marshal(reqBody)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT count\(\*\) FROM "content"."contents" WHERE`).
		WithArgs("article", "duplicate").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`INSERT INTO "content"."contents"`).
		WillReturnError(&pgconn.PgError{Code: "23505"})
	mock.ExpectRollback()

	ctx, rec := testCrudContext(http.MethodPost, "/api/v1/contents", bytes.NewReader(body))
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 10, Role: "admin"})

	c.Create(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusConflict, "content slug is already taken for this type")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
