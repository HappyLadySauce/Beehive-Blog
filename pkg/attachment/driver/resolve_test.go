package driver

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestResolveMountForWriteUsesDefaultMount(t *testing.T) {
	db, mock := newDriverTestDB(t)
	store := NewStore(db)
	registry := testRegistry()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attachment"."storage_mounts"`)).
		WithArgs(true, false, "error", 1).
		WillReturnRows(storageMountRows().AddRow(
			int64(10), "local", "/local", "Local", json.RawMessage(`{"root":"data/attachments"}`),
			0, true, false, "unknown", now, now, nil,
		))

	mount, backend, err := ResolveMountForWrite(context.Background(), store, registry, nil)
	if err != nil {
		t.Fatalf("ResolveMountForWrite: %v", err)
	}
	if mount.ID != 10 || backend.DriverName() != DriverLocal {
		t.Fatalf("resolved mount/backend = %d/%s", mount.ID, backend.DriverName())
	}
}

func TestResolveMountForWriteRejectsDisabledExplicitMount(t *testing.T) {
	db, mock := newDriverTestDB(t)
	store := NewStore(db)
	registry := testRegistry()
	now := time.Now()
	mountID := int64(10)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attachment"."storage_mounts"`)).
		WithArgs(mountID, 1).
		WillReturnRows(storageMountRows().AddRow(
			mountID, "local", "/local", "Local", json.RawMessage(`{"root":"data/attachments"}`),
			0, true, true, "unknown", now, now, nil,
		))

	if _, _, err := ResolveMountForWrite(context.Background(), store, registry, &mountID); err == nil {
		t.Fatalf("ResolveMountForWrite should reject disabled mount")
	}
}

func newDriverTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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
	return db, mock
}

func testRegistry() *DriverRegistry {
	registry := NewDriverRegistry()
	registry.Register(DriverLocal, NewLocalDriver)
	return registry
}

func storageMountRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "driver_name", "mount_path", "name", "config", "order_index", "is_default", "disabled", "status", "created_at", "updated_at", "deleted_at",
	})
}
