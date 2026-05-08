package users

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/httpx"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

func TestRegisterPrecheckUsernameConflict(t *testing.T) {
	c, mock := newRegisterTestController(t)
	expectUsernameLookup(mock, "bob", 9, "bob")
	req := &v1.RegisterRequest{Username: "bob", Password: "password12"}

	_, err := c.Register(testRegisterContext(), req)
	assertAppError(t, err, http.StatusConflict, "username is already taken")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRegisterPrecheckEmailConflict(t *testing.T) {
	c, mock := newRegisterTestController(t)
	expectUsernameLookup(mock, "alice", 0, "")
	expectEmailLookup(mock, "taken@example.com", 3, "other", "taken@example.com")
	req := &v1.RegisterRequest{Username: "alice", Password: "password12", Email: "taken@example.com"}

	_, err := c.Register(testRegisterContext(), req)
	assertAppError(t, err, http.StatusConflict, "email is already registered")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRegisterUniqueViolationOnCreateUsername(t *testing.T) {
	c, mock := newRegisterTestController(t)
	expectUsernameLookup(mock, "alice", 0, "")
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "identity"\."users".*RETURNING "id"`).
		WillReturnError(&pgconn.PgError{Code: "23505", ConstraintName: "ux_identity_users_username"})
	mock.ExpectRollback()

	req := &v1.RegisterRequest{Username: "alice", Password: "password12"}
	_, err := c.Register(testRegisterContext(), req)
	assertAppError(t, err, http.StatusConflict, "username is already taken")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRegisterUniqueViolationOnCreateEmail(t *testing.T) {
	c, mock := newRegisterTestController(t)
	expectUsernameLookup(mock, "alice", 0, "")
	expectEmailLookup(mock, "dup@example.com", 0, "", "")
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "identity"\."users".*RETURNING "id"`).
		WillReturnError(&pgconn.PgError{Code: "23505", ConstraintName: "ux_identity_users_email"})
	mock.ExpectRollback()

	req := &v1.RegisterRequest{Username: "alice", Password: "password12", Email: "dup@example.com"}
	_, err := c.Register(testRegisterContext(), req)
	assertAppError(t, err, http.StatusConflict, "email is already registered")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRegisterUniqueViolationUnknownConstraint(t *testing.T) {
	c, mock := newRegisterTestController(t)
	expectUsernameLookup(mock, "alice", 0, "")
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "identity"\."users".*RETURNING "id"`).
		WillReturnError(&pgconn.PgError{Code: "23505", ConstraintName: "some_other_uq"})
	mock.ExpectRollback()

	req := &v1.RegisterRequest{Username: "alice", Password: "password12"}
	_, err := c.Register(testRegisterContext(), req)
	assertAppError(t, err, http.StatusConflict, "registration conflict")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRegisterCreateUserNonUniqueError(t *testing.T) {
	c, mock := newRegisterTestController(t)
	expectUsernameLookup(mock, "alice", 0, "")
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "identity"\."users".*RETURNING "id"`).
		WillReturnError(gorm.ErrInvalidDB)
	mock.ExpectRollback()

	req := &v1.RegisterRequest{Username: "alice", Password: "password12"}
	_, err := c.Register(testRegisterContext(), req)
	assertAppError(t, err, http.StatusInternalServerError, "failed to register user")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRegisterSuccess(t *testing.T) {
	c, mock := newRegisterTestController(t)
	expectUsernameLookup(mock, "alice", 0, "")
	mock.ExpectBegin()
	expectUserInsert(mock, 42)
	expectCredentialInsert(mock, 7)
	expectSessionInsert(mock, 10)
	expectSessionHashUpdate(mock, 10)
	mock.ExpectCommit()

	req := &v1.RegisterRequest{Username: "alice", Password: "password12"}
	resp, err := c.Register(testRegisterContext(), req)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if resp.Token.AccessToken == "" || resp.Token.RefreshToken == "" {
		t.Fatalf("missing tokens in response: %+v", resp.Token)
	}
	if resp.Token.TokenType != jwt.TokenTypeBearer {
		t.Fatalf("TokenType = %q, want %q", resp.Token.TokenType, jwt.TokenTypeBearer)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRegisterBindInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users/register", strings.NewReader(`{}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	u := NewUsersController(&svc.ServiceContext{})
	httpx.HandleJSON(u.Register)(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("HTTP status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var envelope common.BaseResponse
	if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if envelope.Code != http.StatusBadRequest {
		t.Fatalf("envelope code = %d, want %d", envelope.Code, http.StatusBadRequest)
	}
}

func newRegisterTestController(t *testing.T) (*UsersController, sqlmock.Sqlmock) {
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
		Issuer:     "beehive-blog-register-test",
		Secret:     "0123456789abcdef0123456789abcdef",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	})
	if err != nil {
		t.Fatalf("NewIssuer() error = %v", err)
	}
	return NewUsersController(&svc.ServiceContext{
		DB:    db,
		Token: issuer,
	}), mock
}

func testRegisterContext() *gin.Context {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/register", bytes.NewReader(nil))
	req.RemoteAddr = "203.0.113.11:12345"
	req.Header.Set("User-Agent", "register-test")
	ctx.Request = req
	return ctx
}

func expectUsernameLookup(mock sqlmock.Sqlmock, username string, existingID int64, existingUsername string) {
	q := mock.ExpectQuery(`SELECT .*FROM "identity"\."users".*WHERE username = \$1 AND "users"\."deleted_at" IS NULL.*LIMIT \$2`).
		WithArgs(username, 1)
	if existingID <= 0 {
		q.WillReturnRows(sqlmock.NewRows(userColumns()))
		return
	}
	q.WillReturnRows(sqlmock.NewRows(userColumns()).AddRow(
		existingID, existingUsername, nil, nil, nil, nil, "member", "active", nil,
		time.Now(), time.Now(), nil,
	))
}

func expectEmailLookup(mock sqlmock.Sqlmock, email string, existingID int64, existingUsername, existingEmail string) {
	q := mock.ExpectQuery(`SELECT .*FROM "identity"\."users".*WHERE email = \$1 AND "users"\."deleted_at" IS NULL.*LIMIT \$2`).
		WithArgs(email, 1)
	if existingID <= 0 {
		q.WillReturnRows(sqlmock.NewRows(userColumns()))
		return
	}
	q.WillReturnRows(sqlmock.NewRows(userColumns()).AddRow(
		existingID, existingUsername, existingEmail, nil, nil, nil, "member", "active", nil,
		time.Now(), time.Now(), nil,
	))
}

func expectUserInsert(mock sqlmock.Sqlmock, id int64) {
	mock.ExpectQuery(`INSERT INTO "identity"\."users".*RETURNING "id"`).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
}

func expectCredentialInsert(mock sqlmock.Sqlmock, id int64) {
	mock.ExpectQuery(`INSERT INTO "identity"\."user_credentials".*RETURNING "id"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
}

func expectSessionInsert(mock sqlmock.Sqlmock, newID int64) {
	mock.ExpectQuery(`INSERT INTO "identity"\."user_sessions".*RETURNING "id"`).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(newID))
}

func expectSessionHashUpdate(mock sqlmock.Sqlmock, sessionID int64) {
	mock.ExpectExec(`UPDATE "identity"\."user_sessions" SET .*"refresh_token_hash".*"updated_at".*WHERE "id" = \$3`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sessionID).
		WillReturnResult(sqlmock.NewResult(0, 1))
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
