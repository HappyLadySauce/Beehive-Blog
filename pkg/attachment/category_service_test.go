package attachment

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newCategoryServiceMock(t *testing.T) (*CategoryService, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return NewCategoryService(db), mock
}

func TestCategoryServiceListAdmin(t *testing.T) {
	svc, mock := newCategoryServiceMock(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .* FROM "attachment"\."categories"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "parent_id", "name", "slug", "description", "icon", "path", "depth", "sort_order", "status", "created_at", "updated_at", "deleted_at",
		}).AddRow(int64(1), nil, "Images", "images", nil, nil, "/images", 0, 0, CategoryStatusActive, now, now, nil))

	rows, err := svc.List(context.Background(), Actor{UID: 1, Role: RoleAdmin})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(rows) != 1 || rows[0].Name != "Images" {
		t.Fatalf("rows: %+v", rows)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql: %v", err)
	}
}

func TestCategoryServiceListForbidden(t *testing.T) {
	svc, _ := newCategoryServiceMock(t)
	_, err := svc.List(context.Background(), Actor{UID: 2, Role: "user"})
	if err != ErrForbidden {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestCategoryServiceDeleteNotFound(t *testing.T) {
	svc, mock := newCategoryServiceMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "attachment"\."categories" SET "deleted_at"=\$1 WHERE id = \$2`).
		WithArgs(sqlmock.AnyArg(), int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := svc.Delete(context.Background(), Actor{UID: 1, Role: RoleAdmin}, 99)
	if err != ErrNotFound {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql: %v", err)
	}
}

func TestCategoryServiceCreateValidatesName(t *testing.T) {
	svc, _ := newCategoryServiceMock(t)
	_, err := svc.Create(context.Background(), Actor{UID: 1, Role: RoleAdmin}, CategoryCreateInput{
		Name: "   ",
		Slug: "slug",
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}
}
