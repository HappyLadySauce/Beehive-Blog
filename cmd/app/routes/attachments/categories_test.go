package attachments

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	pkgattachment "github.com/HappyLadySauce/Beehive-Blog/pkg/attachment"
)

func TestListCategoriesForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := mustNewAttachmentsController(t)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachment/categories", nil)
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 2, Role: "user"})
	h.ListCategories(ctx)
	assertEnvelopeCode(t, rec, http.StatusForbidden)
}

func TestGetCategoryInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := mustNewAttachmentsController(t)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachment/categories/abc", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "abc"}}
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
	h.GetCategory(ctx)
	assertEnvelopeCode(t, rec, http.StatusBadRequest)
}

func TestGetCategoryNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock := mustNewAttachmentsControllerWithMock(t)
	mock.ExpectQuery(`SELECT .* FROM "attachment"\."categories" WHERE id = \$1`).
		WithArgs(int64(404), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/attachment/categories/404", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "404"}}
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
	h.GetCategory(ctx)
	assertEnvelopeCode(t, rec, http.StatusNotFound)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestDeleteCategoryNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock := mustNewAttachmentsControllerWithMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "attachment"\."categories" SET "deleted_at"=\$1 WHERE id = \$2`).
		WithArgs(sqlmock.AnyArg(), int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/attachment/categories/99", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "99"}}
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 1, Role: pkgattachment.RoleAdmin})
	h.DeleteCategory(ctx)
	assertEnvelopeCode(t, rec, http.StatusNotFound)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
