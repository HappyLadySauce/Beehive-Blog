package oauth

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestFindOrCreateUserUsesExistingProviderIdentity(t *testing.T) {
	db, mock := newOAuthTestDB(t)
	expectIdentityQuery(mock, "github", "123", 42)
	expectUserByIDQuery(mock, 42, "alice", "changed@example.com")

	user, isNew, err := FindOrCreateUser(db, &GitHubUser{ID: 123, Login: "alice"}, "new@example.com")
	if err != nil {
		t.Fatalf("FindOrCreateUser() error = %v", err)
	}
	if isNew {
		t.Fatalf("isNew = true, want false")
	}
	if user.ID != 42 {
		t.Fatalf("user id = %d, want 42", user.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestFindOrCreateUserBindsExistingEmailUser(t *testing.T) {
	db, mock := newOAuthTestDB(t)
	expectMissingIdentityQuery(mock, "github", "123")
	expectUserByEmailQuery(mock, "alice@example.com", 42, "alice")
	mock.ExpectBegin()
	expectIdentityInsert(mock, 42, "github", "123", "alice@example.com")
	mock.ExpectCommit()

	user, isNew, err := FindOrCreateUser(db, &GitHubUser{ID: 123, Login: "alice"}, "alice@example.com")
	if err != nil {
		t.Fatalf("FindOrCreateUser() error = %v", err)
	}
	if isNew {
		t.Fatalf("isNew = true, want false")
	}
	if user.ID != 42 {
		t.Fatalf("user id = %d, want 42", user.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestFindOrCreateUserCreatesUserAndIdentity(t *testing.T) {
	db, mock := newOAuthTestDB(t)
	expectMissingIdentityQuery(mock, "github", "123")
	expectMissingUserByEmailQuery(mock, "alice@example.com")
	mock.ExpectBegin()
	expectUserInsert(mock, 42)
	expectIdentityInsert(mock, 42, "github", "123", "alice@example.com")
	mock.ExpectCommit()

	user, isNew, err := FindOrCreateUser(db, &GitHubUser{ID: 123, Login: "alice", Name: "Alice"}, "alice@example.com")
	if err != nil {
		t.Fatalf("FindOrCreateUser() error = %v", err)
	}
	if !isNew {
		t.Fatalf("isNew = false, want true")
	}
	if user.ID != 42 {
		t.Fatalf("user id = %d, want 42", user.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func newOAuthTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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
	return db, mock
}

func expectIdentityQuery(mock sqlmock.Sqlmock, provider, subject string, userID int64) {
	mock.ExpectQuery(`SELECT .*FROM "identity"\."user_identities".*WHERE .*provider = \$1 AND provider_subject = \$2.*LIMIT \$3`).
		WithArgs(provider, subject, 1).
		WillReturnRows(sqlmock.NewRows(identityColumns()).
			AddRow(9, userID, provider, subject, nil, time.Now(), time.Now(), nil))
}

func expectMissingIdentityQuery(mock sqlmock.Sqlmock, provider, subject string) {
	mock.ExpectQuery(`SELECT .*FROM "identity"\."user_identities".*WHERE .*provider = \$1 AND provider_subject = \$2.*LIMIT \$3`).
		WithArgs(provider, subject, 1).
		WillReturnRows(sqlmock.NewRows(identityColumns()))
}

func expectUserByIDQuery(mock sqlmock.Sqlmock, id int64, username, email string) {
	mock.ExpectQuery(`SELECT .*FROM "identity"\."users".*WHERE "users"\."id" = \$1 AND "users"\."deleted_at" IS NULL.*LIMIT \$2`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows(userColumns()).AddRow(id, username, email, nil, nil, nil, "member", "active", nil, time.Now(), time.Now(), nil))
}

func expectUserByEmailQuery(mock sqlmock.Sqlmock, email string, id int64, username string) {
	mock.ExpectQuery(`SELECT .*FROM "identity"\."users".*WHERE email = \$1 AND "users"\."deleted_at" IS NULL.*LIMIT \$2`).
		WithArgs(email, 1).
		WillReturnRows(sqlmock.NewRows(userColumns()).AddRow(id, username, email, nil, nil, nil, "member", "active", nil, time.Now(), time.Now(), nil))
}

func expectMissingUserByEmailQuery(mock sqlmock.Sqlmock, email string) {
	mock.ExpectQuery(`SELECT .*FROM "identity"\."users".*WHERE email = \$1 AND "users"\."deleted_at" IS NULL.*LIMIT \$2`).
		WithArgs(email, 1).
		WillReturnRows(sqlmock.NewRows(userColumns()))
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

func expectIdentityInsert(mock sqlmock.Sqlmock, userID int64, provider, subject, email string) {
	mock.ExpectQuery(`INSERT INTO "identity"\."user_identities".*RETURNING "id"`).
		WithArgs(userID, provider, subject, email, sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
}

func identityColumns() []string {
	return []string{"id", "user_id", "provider", "provider_subject", "email_at_bind", "created_at", "updated_at", "deleted_at"}
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
