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

	mock.ExpectQuery(`SELECT count\(\*\) FROM "attachment"\."categories"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))
	mock.ExpectQuery(`SELECT .* FROM "attachment"\."categories"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "parent_id", "name", "slug", "description", "icon", "path", "depth", "sort_order", "status", "created_at", "updated_at", "deleted_at",
		}).AddRow(int64(1), nil, "Images", "images", nil, nil, "/images", 0, 0, CategoryStatusActive, now, now, nil))

	rows, total, err := svc.List(context.Background(), Actor{UID: 1, Role: RoleAdmin}, 1, 20, "", "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 5 {
		t.Fatalf("total = %d, want 5", total)
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
	_, _, err := svc.List(context.Background(), Actor{UID: 2, Role: "user"}, 1, 20, "", "")
	if err != ErrForbidden {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestCategoryServiceListWithStatusFilter(t *testing.T) {
	svc, mock := newCategoryServiceMock(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "attachment"\."categories" WHERE status = \$1`).
		WithArgs("disabled").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT \* FROM "attachment"\."categories" WHERE status = \$1.*ORDER BY path ASC LIMIT \$2`).
		WithArgs("disabled", 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "parent_id", "name", "slug", "description", "icon", "path", "depth", "sort_order", "status", "created_at", "updated_at", "deleted_at",
		}).AddRow(int64(2), nil, "Archived", "archived", nil, nil, "/archived", 0, 0, "disabled", now, now, nil))

	rows, total, err := svc.List(context.Background(), Actor{UID: 1, Role: RoleAdmin}, 1, 20, "disabled", "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 {
		t.Fatalf("total = %d, want 1", total)
	}
	if len(rows) != 1 || rows[0].Status != "disabled" {
		t.Fatalf("rows: %+v", rows)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql: %v", err)
	}
}

func TestCategoryServiceListWithSearch(t *testing.T) {
	svc, mock := newCategoryServiceMock(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "attachment"\."categories" WHERE \(name ILIKE \$1 OR slug ILIKE \$2\)`).
		WithArgs("%img%", "%img%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT \* FROM "attachment"\."categories" WHERE \(name ILIKE \$1 OR slug ILIKE \$2\).*ORDER BY path ASC LIMIT \$3`).
		WithArgs("%img%", "%img%", 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "parent_id", "name", "slug", "description", "icon", "path", "depth", "sort_order", "status", "created_at", "updated_at", "deleted_at",
		}).AddRow(int64(1), nil, "Images", "images", nil, nil, "/images", 0, 0, CategoryStatusActive, now, now, nil))

	rows, total, err := svc.List(context.Background(), Actor{UID: 1, Role: RoleAdmin}, 1, 20, "", "img")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 {
		t.Fatalf("total = %d, want 1", total)
	}
	if len(rows) != 1 || rows[0].Name != "Images" {
		t.Fatalf("rows: %+v", rows)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql: %v", err)
	}
}

func TestCategoryServiceListPagination(t *testing.T) {
	svc, mock := newCategoryServiceMock(t)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "attachment"\."categories"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(25)))
	mock.ExpectQuery(`SELECT \* FROM "attachment"\."categories" WHERE.*ORDER BY path ASC LIMIT \$1 OFFSET \$2`).
		WithArgs(5, 5).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "parent_id", "name", "slug", "description", "icon", "path", "depth", "sort_order", "status", "created_at", "updated_at", "deleted_at",
		}))

	rows, total, err := svc.List(context.Background(), Actor{UID: 1, Role: RoleAdmin}, 2, 5, "", "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 25 {
		t.Fatalf("total = %d, want 25", total)
	}
	if len(rows) != 0 {
		t.Fatalf("rows = %d, want 0", len(rows))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql: %v", err)
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
