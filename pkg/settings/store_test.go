package settings_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/settings"
	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

func TestEnsureSingletonInsertsWhenMissing(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(regexp.MustCompile(`SELECT count\(\*\) FROM "setting"\."application_settings"`).String()).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.MustCompile(`INSERT INTO "setting"\."application_settings"`).String()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int16(1)))
	mock.ExpectCommit()

	store := settings.NewStore(db)
	seed := settingtypes.DefaultApplicationSettings()
	if err := store.EnsureSingleton(context.Background(), seed); err != nil {
		t.Fatalf("EnsureSingleton: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureSingletonNoInsertWhenRowExists(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(regexp.MustCompile(`SELECT count\(\*\) FROM "setting"\."application_settings"`).String()).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	store := settings.NewStore(db)
	if err := store.EnsureSingleton(context.Background(), settingtypes.DefaultApplicationSettings()); err != nil {
		t.Fatalf("EnsureSingleton: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureSingletonTreatsUniqueViolationAsSuccess(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(regexp.MustCompile(`SELECT count\(\*\) FROM "setting"\."application_settings"`).String()).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.MustCompile(`INSERT INTO "setting"\."application_settings"`).String()).
		WillReturnError(&pgconn.PgError{Code: "23505"})
	mock.ExpectRollback()

	store := settings.NewStore(db)
	if err := store.EnsureSingleton(context.Background(), settingtypes.DefaultApplicationSettings()); err != nil {
		t.Fatalf("EnsureSingleton: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureSingletonInvalidPayloadRejected(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	store := settings.NewStore(db)
	bad := settingtypes.DefaultApplicationSettings()
	bad.Email.TLS = "bad"
	if err := store.EnsureSingleton(context.Background(), bad); err == nil {
		t.Fatal("expected validation error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// Ensure model table name is wired for GORM queries.
// 确保 GORM 使用正确的表名。
func TestApplicationSettingTableName(t *testing.T) {
	var m model.ApplicationSetting
	if m.TableName() != "setting.application_settings" {
		t.Fatalf("TableName = %q", m.TableName())
	}
}
