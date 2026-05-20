package tags

import (
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
)

func TestMapTagCreateUniqueViolation(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "23505"}
	err := mapTagCreateUniqueViolation(pgErr)

	appErr, ok := err.(*common.AppError)
	if !ok {
		t.Fatalf("expected *common.AppError, got %T", err)
	}
	if appErr.HTTPStatus != http.StatusConflict {
		t.Fatalf("HTTP status = %d, want 409", appErr.HTTPStatus)
	}
	if appErr.Message != "tag slug is already taken" {
		t.Fatalf("message = %q, want 'tag slug is already taken'", appErr.Message)
	}
}

func TestMapTagCreateUniqueViolationNonPgError(t *testing.T) {
	err := mapTagCreateUniqueViolation(sqlmock.ErrCancelled)

	appErr, ok := err.(*common.AppError)
	if !ok {
		t.Fatalf("expected *common.AppError, got %T", err)
	}
	if appErr.HTTPStatus != http.StatusInternalServerError {
		t.Fatalf("HTTP status = %d, want 500", appErr.HTTPStatus)
	}
}
