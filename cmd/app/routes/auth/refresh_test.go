package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	routeauth "github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/auth"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	authsession "github.com/HappyLadySauce/Beehive-Blog/pkg/auth/session"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

func TestRefreshRotatesSessionAndUsesCurrentDatabaseRole(t *testing.T) {
	controller, mock, issuer := newRefreshTestController(t)
	pair, err := issuer.IssueSessionPair(42, "admin", 7, "old-jti")
	if err != nil {
		t.Fatalf("IssueSessionPair() error = %v", err)
	}

	mock.ExpectBegin()
	expectSessionQuery(mock, sessionRow{
		id:         7,
		userID:     42,
		hash:       authsession.HashRefreshToken(pair.Refresh.Token),
		jti:        "old-jti",
		expiresAt:  time.Now().Add(time.Hour),
		revokedAt:  nil,
		rotatedAt:  nil,
		createdAt:  time.Now(),
		updatedAt:  time.Now(),
		createdIP:  "203.0.113.1",
		userAgent:  "old-agent",
		lastUsedAt: nil,
	})
	expectUserQuery(mock, userRow{id: 42, username: "alice", role: "member", status: "active"})
	expectRotateUpdate(mock, 7)
	expectSessionInsert(mock, 8)
	expectSessionHashUpdate(mock, 8)
	mock.ExpectCommit()

	rec, envelope := performRefresh(t, controller, pair.Refresh.Token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	accessClaims, err := issuer.ParseAccess(envelope.Data.Token.AccessToken)
	if err != nil {
		t.Fatalf("ParseAccess() error = %v", err)
	}
	if accessClaims.Role != "member" {
		t.Fatalf("access role = %q, want member", accessClaims.Role)
	}
	refreshClaims, err := issuer.ParseRefresh(envelope.Data.Token.RefreshToken)
	if err != nil {
		t.Fatalf("ParseRefresh() error = %v", err)
	}
	if refreshClaims.SID != 8 {
		t.Fatalf("new refresh sid = %d, want 8", refreshClaims.SID)
	}
	if envelope.Data.Token.RefreshToken == pair.Refresh.Token {
		t.Fatalf("refresh token was not rotated")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRefreshRejectsDisabledUser(t *testing.T) {
	controller, mock, issuer := newRefreshTestController(t)
	pair, err := issuer.IssueSessionPair(42, "member", 7, "old-jti")
	if err != nil {
		t.Fatalf("IssueSessionPair() error = %v", err)
	}

	mock.ExpectBegin()
	expectSessionQuery(mock, sessionRow{
		id:        7,
		userID:    42,
		hash:      authsession.HashRefreshToken(pair.Refresh.Token),
		jti:       "old-jti",
		expiresAt: time.Now().Add(time.Hour),
		createdAt: time.Now(),
		updatedAt: time.Now(),
	})
	expectUserQuery(mock, userRow{id: 42, username: "alice", role: "member", status: "disabled"})
	mock.ExpectRollback()

	rec, envelope := performRefresh(t, controller, pair.Refresh.Token)
	assertRefreshHTTPError(t, rec, envelope, http.StatusForbidden, "account is not allowed to login")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRefreshRejectsReusedRotatedToken(t *testing.T) {
	controller, mock, issuer := newRefreshTestController(t)
	pair, err := issuer.IssueSessionPair(42, "member", 7, "old-jti")
	if err != nil {
		t.Fatalf("IssueSessionPair() error = %v", err)
	}
	rotatedAt := time.Now().Add(-time.Minute)

	mock.ExpectBegin()
	expectSessionQuery(mock, sessionRow{
		id:        7,
		userID:    42,
		hash:      authsession.HashRefreshToken(pair.Refresh.Token),
		jti:       "old-jti",
		expiresAt: time.Now().Add(time.Hour),
		rotatedAt: &rotatedAt,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	})
	expectRevokeUpdate(mock)
	mock.ExpectCommit()

	rec, envelope := performRefresh(t, controller, pair.Refresh.Token)
	assertRefreshHTTPError(t, rec, envelope, http.StatusUnauthorized, "invalid or expired refresh token")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRefreshRejectsAccessTokenUse(t *testing.T) {
	controller, mock, issuer := newRefreshTestController(t)
	pair, err := issuer.IssueSessionPair(42, "member", 7, "old-jti")
	if err != nil {
		t.Fatalf("IssueSessionPair() error = %v", err)
	}

	rec, envelope := performRefresh(t, controller, pair.Access.Token)
	assertRefreshHTTPError(t, rec, envelope, http.StatusUnauthorized, "invalid or expired refresh token")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRefreshRejectsJTIOrHashMismatch(t *testing.T) {
	controller, mock, issuer := newRefreshTestController(t)
	pair, err := issuer.IssueSessionPair(42, "member", 7, "old-jti")
	if err != nil {
		t.Fatalf("IssueSessionPair() error = %v", err)
	}

	mock.ExpectBegin()
	expectSessionQuery(mock, sessionRow{
		id:         7,
		userID:     42,
		hash:       "different-hash-not-matching",
		jti:        "different-jti",
		expiresAt:  time.Now().Add(time.Hour),
		revokedAt:  nil,
		rotatedAt:  nil,
		createdAt:  time.Now(),
		updatedAt:  time.Now(),
		createdIP:  "203.0.113.1",
		userAgent:  "old-agent",
		lastUsedAt: nil,
	})
	expectRevokeUpdate(mock)
	mock.ExpectCommit()

	rec, envelope := performRefresh(t, controller, pair.Refresh.Token)
	assertRefreshHTTPError(t, rec, envelope, http.StatusUnauthorized, "invalid or expired refresh token")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

type sessionRow struct {
	id         int64
	userID     int64
	hash       string
	jti        string
	expiresAt  time.Time
	revokedAt  *time.Time
	rotatedAt  *time.Time
	createdIP  string
	userAgent  string
	lastUsedAt *time.Time
	createdAt  time.Time
	updatedAt  time.Time
}

type userRow struct {
	id       int64
	username string
	role     string
	status   string
}

func newRefreshTestController(t *testing.T) (*routeauth.AuthController, sqlmock.Sqlmock, *jwt.Issuer) {
	t.Helper()
	gin.SetMode(gin.TestMode)
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
	return routeauth.NewAuthController(
		&svc.ServiceContext{
			DB:    db,
			Token: issuer,
		},
	), mock, issuer
}

func performRefresh(t *testing.T, controller *routeauth.AuthController, refreshToken string) (*httptest.ResponseRecorder, refreshEnvelope) {
	t.Helper()
	body, err := json.Marshal(v1.RefreshRequest{RefreshToken: refreshToken})
	if err != nil {
		t.Fatalf("Marshal refresh request: %v", err)
	}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "203.0.113.10:12345"
	req.Header.Set("User-Agent", "refresh-test")
	ctx.Request = req
	controller.Refresh(ctx)

	var envelope refreshEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("Unmarshal refresh response %q: %v", rec.Body.String(), err)
	}
	return rec, envelope
}

func expectSessionQuery(mock sqlmock.Sqlmock, row sessionRow) {
	mock.ExpectQuery(`SELECT .*FROM "identity"\."user_sessions".*WHERE id = \$1 AND user_id = \$2.*LIMIT \$3 FOR UPDATE`).
		WithArgs(row.id, row.userID, 1).
		WillReturnRows(sqlmock.NewRows(sessionColumns()).
			AddRow(row.id, row.userID, row.hash, row.jti, row.expiresAt, row.revokedAt, nil, row.rotatedAt, row.createdIP, row.userAgent, row.lastUsedAt, row.createdAt, row.updatedAt))
}

func expectUserQuery(mock sqlmock.Sqlmock, row userRow) {
	mock.ExpectQuery(`SELECT .*FROM "identity"\."users".*WHERE "users"\."id" = \$1 AND "users"\."deleted_at" IS NULL.*LIMIT \$2`).
		WithArgs(row.id, 1).
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(row.id, row.username, nil, nil, nil, nil, row.role, row.status, nil, time.Now(), time.Now(), nil))
}

func expectRotateUpdate(mock sqlmock.Sqlmock, sessionID int64) {
	mock.ExpectExec(`UPDATE "identity"\."user_sessions" SET .*"last_used_at".*"rotated_at".*"updated_at".*WHERE "id" = \$4`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sessionID).
		WillReturnResult(sqlmock.NewResult(0, 1))
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

func expectRevokeUpdate(mock sqlmock.Sqlmock) {
	mock.ExpectExec(`UPDATE "identity"\."user_sessions" SET .*"revoked_at".*"revoked_reason".*"updated_at".*WHERE .*revoked_at IS NULL`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

func sessionColumns() []string {
	return []string{
		"id",
		"user_id",
		"refresh_token_hash",
		"refresh_jti",
		"expires_at",
		"revoked_at",
		"revoked_reason",
		"rotated_at",
		"created_ip",
		"user_agent",
		"last_used_at",
		"created_at",
		"updated_at",
	}
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

type refreshEnvelope struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    v1.RefreshResponse `json:"data"`
}

func assertRefreshHTTPError(t *testing.T, rec *httptest.ResponseRecorder, envelope refreshEnvelope, status int, message string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, status, rec.Body.String())
	}
	if envelope.Message != message {
		t.Fatalf("Message = %q, want %q", envelope.Message, message)
	}
}
