package settings

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestProviderRefreshLoadsSnapshot(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte(`{"email":{"enabled":false,"host":"","port":587,"username":"","password":"","from":"","from_name":"","tls":"starttls"}}`)
	now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`SELECT .* FROM "setting"."application_settings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "revision", "payload", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 2, payload, now, now, nil))

	p := NewProvider(NewStore(db))
	if err := p.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if p.CachedRevision() != 2 {
		t.Fatalf("revision = %d", p.CachedRevision())
	}
	cur := p.Current()
	if cur.Email.Port != 587 {
		t.Fatalf("port = %d", cur.Email.Port)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProviderSaveAndRefreshUpdatesSnapshotWithoutReload(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	payload := []byte(`{"email":{"enabled":false,"host":"","port":587,"username":"","password":"","from":"","from_name":"","tls":"starttls"}}`)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM "setting"."application_settings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "revision", "payload", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 2, payload, now, now, nil))
	mock.ExpectExec(`UPDATE "setting"."application_settings"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	p := NewProvider(NewStore(db))
	next := p.Current()
	next.Email.Host = "smtp.example.com"
	if err := p.SaveAndRefresh(context.Background(), next); err != nil {
		t.Fatalf("SaveAndRefresh: %v", err)
	}
	if p.CachedRevision() != 3 {
		t.Fatalf("revision = %d", p.CachedRevision())
	}
	if got := p.Current().Email.Host; got != "smtp.example.com" {
		t.Fatalf("host = %q", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
