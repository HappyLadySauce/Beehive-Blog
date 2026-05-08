package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

func TestRefreshUsesCurrentDatabaseRole(t *testing.T) {
	controller, mock, issuer := newRefreshTestController(t)
	pair, err := issuer.IssuePair(42, "admin")
	if err != nil {
		t.Fatalf("IssuePair() error = %v", err)
	}
	expectUserQuery(mock, userRow{
		id:       42,
		username: "alice",
		role:     "member",
		status:   "active",
	})

	resp, err := controller.Refresh(nil, &v1.RefreshRequest{RefreshToken: pair.Refresh.Token})
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	claims, err := issuer.ParseAccess(resp.Token.AccessToken)
	if err != nil {
		t.Fatalf("ParseAccess() error = %v", err)
	}
	if claims.Role != "member" {
		t.Fatalf("access role = %q, want member", claims.Role)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRefreshRejectsDisabledUser(t *testing.T) {
	controller, mock, issuer := newRefreshTestController(t)
	pair, err := issuer.IssuePair(42, "member")
	if err != nil {
		t.Fatalf("IssuePair() error = %v", err)
	}
	expectUserQuery(mock, userRow{
		id:       42,
		username: "alice",
		role:     "member",
		status:   "disabled",
	})

	_, err = controller.Refresh(nil, &v1.RefreshRequest{RefreshToken: pair.Refresh.Token})
	if err == nil {
		t.Fatalf("Refresh() error = nil, want error")
	}
	assertAppError(t, err, http.StatusForbidden, "account is not allowed to login")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRefreshRejectsMissingUser(t *testing.T) {
	controller, mock, issuer := newRefreshTestController(t)
	pair, err := issuer.IssuePair(42, "member")
	if err != nil {
		t.Fatalf("IssuePair() error = %v", err)
	}
	expectMissingUserQuery(mock)

	_, err = controller.Refresh(nil, &v1.RefreshRequest{RefreshToken: pair.Refresh.Token})
	if err == nil {
		t.Fatalf("Refresh() error = nil, want error")
	}
	assertAppError(t, err, http.StatusUnauthorized, "invalid or expired refresh token")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRefreshRejectsAccessTokenUse(t *testing.T) {
	controller, mock, issuer := newRefreshTestController(t)
	pair, err := issuer.IssuePair(42, "member")
	if err != nil {
		t.Fatalf("IssuePair() error = %v", err)
	}

	_, err = controller.Refresh(nil, &v1.RefreshRequest{RefreshToken: pair.Access.Token})
	if err == nil {
		t.Fatalf("Refresh() error = nil, want error")
	}
	assertAppError(t, err, http.StatusUnauthorized, "invalid or expired refresh token")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

type userRow struct {
	id       int64
	username string
	role     string
	status   string
}

func newRefreshTestController(t *testing.T) (*AuthController, sqlmock.Sqlmock, *jwt.Issuer) {
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
	issuer, err := jwt.NewIssuer(&options.JWTOptions{
		Issuer:     "beehive-blog-test",
		Secret:     "0123456789abcdef0123456789abcdef",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	})
	if err != nil {
		t.Fatalf("NewIssuer() error = %v", err)
	}
	return &AuthController{
		svc: svc.ServiceContext{
			DB:    db,
			Token: issuer,
		},
	}, mock, issuer
}

func expectUserQuery(mock sqlmock.Sqlmock, row userRow) {
	mock.ExpectQuery(userQueryPattern()).
		WithArgs(row.id, 1).
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(row.id, row.username, nil, nil, nil, nil, row.role, row.status, nil, time.Now(), time.Now(), nil))
}

func expectMissingUserQuery(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(userQueryPattern()).
		WithArgs(int64(42), 1).
		WillReturnRows(sqlmock.NewRows(userColumns()))
}

func userQueryPattern() string {
	return `SELECT .*FROM "identity"\."users".*WHERE "users"\."id" = \$1 AND "users"\."deleted_at" IS NULL.*LIMIT \$2`
}

func userColumns() []string {
	return []string{
		"id",
		"username",
		"email",
		"nickname",
		"phone",
		"avatar_attachment_id",
		"role",
		"status",
		"last_login_at",
		"created_at",
		"updated_at",
		"deleted_at",
	}
}

func assertAppError(t *testing.T, err error, status int, message string) {
	t.Helper()
	appErr, ok := err.(*common.AppError)
	if !ok {
		t.Fatalf("error type = %T, want *common.AppError", err)
	}
	if appErr.HTTPStatus != status {
		t.Fatalf("HTTPStatus = %d, want %d", appErr.HTTPStatus, status)
	}
	if appErr.Message != message {
		t.Fatalf("Message = %q, want %q", appErr.Message, message)
	}
}
