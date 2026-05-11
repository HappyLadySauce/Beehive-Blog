package settings_test

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"regexp"
	"testing"
	"time"

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

func TestPatchMergesAgainstLockedLatestPayload(t *testing.T) {
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
	payload := []byte(`{"email":{"enabled":false,"host":"old.example.com","port":587,"username":"robot","password":"secret","from":"","from_name":"","tls":"starttls"}}`)
	host := "smtp.example.com"
	patch := &settingtypes.SettingsPatchRequest{
		Email: &settingtypes.EmailSMTPPatch{Host: &host},
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.MustCompile(`SELECT .* FROM "setting"\."application_settings".*FOR UPDATE`).String()).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "revision", "payload", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 2, payload, now, now, nil))
	mock.ExpectExec(regexp.MustCompile(`UPDATE "setting"\."application_settings"`).String()).
		WithArgs(jsonPayloadHas(t, map[string]string{
			"email.host":     "smtp.example.com",
			"email.password": "secret",
		}), int64(3), sqlmock.AnyArg(), int16(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	store := settings.NewStore(db)
	next, rev, err := store.Patch(context.Background(), patch)
	if err != nil {
		t.Fatalf("Patch: %v", err)
	}
	if rev != 3 {
		t.Fatalf("revision = %d, want 3", rev)
	}
	if next.Email.Host != "smtp.example.com" || next.Email.Password != "secret" {
		t.Fatalf("merged email = %+v", next.Email)
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

type jsonPayloadMatcher struct {
	t    *testing.T
	want map[string]string
}

func jsonPayloadHas(t *testing.T, want map[string]string) sqlmock.Argument {
	t.Helper()
	return jsonPayloadMatcher{t: t, want: want}
}

func (m jsonPayloadMatcher) Match(v driver.Value) bool {
	raw, ok := v.([]byte)
	if !ok {
		if s, ok := v.(string); ok {
			raw = []byte(s)
		} else {
			m.t.Logf("payload arg type = %T, want []byte or string", v)
			return false
		}
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		m.t.Logf("payload is not JSON: %v", err)
		return false
	}
	email, ok := doc["email"].(map[string]any)
	if !ok {
		m.t.Logf("payload.email missing or invalid: %v", doc["email"])
		return false
	}
	for path, want := range m.want {
		var got any
		switch path {
		case "email.host":
			got = email["host"]
		case "email.password":
			got = email["password"]
		default:
			m.t.Logf("unsupported json path %q", path)
			return false
		}
		if got != want {
			m.t.Logf("%s = %v, want %v", path, got, want)
			return false
		}
	}
	return true
}
