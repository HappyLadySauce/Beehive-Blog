package filedrivers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/attachment/driver"
)

func TestWriteFileDriverErrorUniqueViolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	pgErr := &pgconn.PgError{Code: "23505"}
	writeFileDriverError(ctx, pgErr)
	if rec.Code != http.StatusConflict {
		t.Fatalf("HTTP = %d, want 409", rec.Code)
	}
}

func TestWriteFileDriverErrorUnsupportedDriver(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	writeFileDriverError(ctx, driver.ErrUnsupportedDriver)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("HTTP = %d, want 400", rec.Code)
	}
}

func TestWriteFileDriverErrorInternal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	writeFileDriverError(ctx, sqlmock.ErrCancelled)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("HTTP = %d, want 500", rec.Code)
	}
	_ = common.NewInternal("internal error", sqlmock.ErrCancelled)
}
