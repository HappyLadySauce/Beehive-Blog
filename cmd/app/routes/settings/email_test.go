package settings_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	routesettings "github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/settings"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/config"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

func TestToResponsePasswordSet(t *testing.T) {
	s := settingtypes.DefaultApplicationSettings()
	s.Email.Password = "secret"
	c, mock := newSettingsTestController(t, s, 7)
	rec, envelope := performGetEmailSettings(t, c)
	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !envelope.Data.Email.PasswordSet {
		t.Fatal("expected PasswordSet true")
	}
	if envelope.Data.Revision != 7 {
		t.Fatalf("revision = %d", envelope.Data.Revision)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestNewSettingsControllerRejectsMissingDependencies(t *testing.T) {
	if _, err := routesettings.NewSettingsController(context.Background(), nil); err == nil || !strings.Contains(err.Error(), "service context is nil") {
		t.Fatalf("NewSettingsController(nil) error = %v, want service context error", err)
	}

	_, err := routesettings.NewSettingsController(context.Background(), &svc.ServiceContext{Config: &config.Config{Email: options.NewEmailSMTPOptions()}})
	if err == nil || !strings.Contains(err.Error(), "database handle is nil") {
		t.Fatalf("NewSettingsController without DB error = %v, want database handle error", err)
	}
}

type settingsEnvelope struct {
	Code int                 `json:"code"`
	Msg  string              `json:"message"`
	Data v1.SettingsResponse `json:"data"`
}

func newSettingsTestController(t *testing.T, s settingtypes.ApplicationSettings, revision int64) (*routesettings.SettingsController, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}

	payload, err := settingtypes.MarshalPayload(s)
	if err != nil {
		t.Fatalf("marshal settings payload: %v", err)
	}
	now := time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.MustCompile(`SELECT count\(\*\) FROM "setting"\."application_settings"`).String()).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT .* FROM "setting"."application_settings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "revision", "payload", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, revision, payload, now, now, nil))

	controller, err := routesettings.NewSettingsController(context.Background(), &svc.ServiceContext{
		Config: &config.Config{Email: options.NewEmailSMTPOptions()},
		DB:     db,
		Token:  &jwt.Issuer{},
	})
	if err != nil {
		t.Fatalf("NewSettingsController: %v", err)
	}
	return controller, mock
}

func performGetEmailSettings(t *testing.T, c *routesettings.SettingsController) (*httptest.ResponseRecorder, settingsEnvelope) {
	t.Helper()
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/settings/email", nil)
	c.GetEmailSettings(ctx)

	var envelope settingsEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return rec, envelope
}
