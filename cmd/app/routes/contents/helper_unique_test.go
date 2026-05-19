package contents

import (
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
)

func TestMapContentCreateUniqueViolation(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "23505"}
	err := mapContentCreateUniqueViolation(pgErr)
	appErr, ok := err.(*common.AppError)
	if !ok {
		t.Fatalf("expected *common.AppError, got %T", err)
	}
	if appErr.HTTPStatus != http.StatusConflict {
		t.Fatalf("HTTP status = %d, want 409", appErr.HTTPStatus)
	}
	if appErr.Message != "content slug is already taken for this type" {
		t.Fatalf("message = %q", appErr.Message)
	}
}

func TestMapContentUpdateUniqueViolation(t *testing.T) {
	err := mapContentUpdateUniqueViolation(sqlmock.ErrCancelled)
	appErr, ok := err.(*common.AppError)
	if !ok {
		t.Fatalf("expected *common.AppError, got %T", err)
	}
	if appErr.HTTPStatus != http.StatusInternalServerError {
		t.Fatalf("HTTP status = %d, want 500", appErr.HTTPStatus)
	}
	if appErr.Message != "failed to update content" {
		t.Fatalf("message = %q", appErr.Message)
	}
}
