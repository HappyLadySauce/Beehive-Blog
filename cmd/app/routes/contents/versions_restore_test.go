package contents

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
)

func versionColumns() []string {
	return []string{"id", "content_id", "version_number", "snapshot_type", "name", "title", "body", "excerpt",
		"change_summary", "created_by", "created_at"}
}

func TestRestoreVersionContentNotFound(t *testing.T) {
	c, mock := newCrudTestController(t)
	body, _ := json.Marshal(v1.RestoreVersionRequest{})

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WillReturnRows(sqlmock.NewRows(contentColumns()))
	mock.ExpectRollback()

	ctx, rec := testCrudContextWithID(http.MethodPost, "/api/v1/contents/999/versions/1/restore", bytes.NewReader(body), "999")
	ctx.Params = append(ctx.Params, gin.Param{Key: "versionNumber", Value: "1"})
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 10, Role: "admin"})

	c.RestoreVersion(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusNotFound, "content not found")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRestoreVersionVersionNotFound(t *testing.T) {
	c, mock := newCrudTestController(t)
	body, _ := json.Marshal(v1.RestoreVersionRequest{})

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "Test", "test", nil, nil, nil, 10, "draft", "public", "allowed", nil, 0, 0, json.RawMessage("{}"), 0, now, now, nil))
	mock.ExpectQuery(`SELECT \* FROM "content"."content_versions" WHERE`).
		WillReturnRows(sqlmock.NewRows(versionColumns()))
	mock.ExpectRollback()

	ctx, rec := testCrudContextWithID(http.MethodPost, "/api/v1/contents/1/versions/99/restore", bytes.NewReader(body), "1")
	ctx.Params = append(ctx.Params, gin.Param{Key: "versionNumber", Value: "99"})
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 10, Role: "admin"})

	c.RestoreVersion(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusNotFound, "version not found")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRestoreVersionSuccess(t *testing.T) {
	c, mock := newCrudTestController(t)
	body, _ := json.Marshal(v1.RestoreVersionRequest{})

	now := time.Now()
	oldBody := "Old body text with several words for testing"
	newBody := "New body text that was written after the version snapshot"

	// 1. Lock content row.
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "Current Title", "test", nil, &newBody, nil, 10, "draft", "public", "allowed", nil, 0, 0, json.RawMessage("{}"), 0, now, now, nil))

	// 2. Load target version.
	mock.ExpectQuery(`SELECT \* FROM "content"."content_versions" WHERE`).
		WithArgs(int64(1), 1, 1).
		WillReturnRows(sqlmock.NewRows(versionColumns()).
			AddRow(1, 1, 1, "manual", "Initial draft", "Old Title", &oldBody, nil, nil, 10, now))

	// 3. Auto snapshot: read max version number.
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(version_number\), 0\) FROM "content"."content_versions" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(1))

	// 4. Create auto snapshot.
	mock.ExpectQuery(`INSERT INTO "content"."content_versions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	// 5. Update content row with version fields.
	mock.ExpectExec(`UPDATE "content"."contents" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	// 6. Post-commit: get() admin to assemble ContentDetailResponse.
	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "Old Title", "test", nil, &oldBody, nil, 10, "draft", "public", "allowed", nil, 8, 1, json.RawMessage("{}"), 0, now, now, nil))
	mock.ExpectQuery(`SELECT "username" FROM "identity"."users" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("admin"))
	// Junction table — no tags.
	mock.ExpectQuery(`SELECT \* FROM "content"."content_tags" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"content_id", "tag_id", "created_at"}))
	// Junction table — no categories.
	mock.ExpectQuery(`SELECT \* FROM "content"."content_categories" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"content_id", "category_id", "created_at"}))

	ctx, rec := testCrudContextWithID(http.MethodPost, "/api/v1/contents/1/versions/1/restore", bytes.NewReader(body), "1")
	ctx.Params = append(ctx.Params, gin.Param{Key: "versionNumber", Value: "1"})
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 10, Role: "admin"})

	c.RestoreVersion(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.ContentDetailResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if resp.Title != "Old Title" {
		t.Fatalf("title = %q, want %q", resp.Title, "Old Title")
	}
	if resp.WordCount != 8 {
		t.Fatalf("word_count = %d, want 8", resp.WordCount)
	}
	if resp.ReadingTimeMinutes != 1 {
		t.Fatalf("reading_time_minutes = %d, want 1", resp.ReadingTimeMinutes)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRestoreVersionCustomChangeSummary(t *testing.T) {
	c, mock := newCrudTestController(t)
	customSummary := "restore test summary"
	body, _ := json.Marshal(v1.RestoreVersionRequest{ChangeSummary: &customSummary})

	now := time.Now()
	oldBody := "Old body"
	newBody := "New body"

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "Current", "test", nil, &newBody, nil, 10, "draft", "public", "allowed", nil, 0, 0, json.RawMessage("{}"), 0, now, now, nil))

	mock.ExpectQuery(`SELECT \* FROM "content"."content_versions" WHERE`).
		WithArgs(int64(1), 1, 1).
		WillReturnRows(sqlmock.NewRows(versionColumns()).
			AddRow(1, 1, 1, "manual", "Initial draft", "Old", &oldBody, nil, nil, 10, now))

	mock.ExpectQuery(`SELECT COALESCE\(MAX\(version_number\), 0\) FROM "content"."content_versions" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(1))

	mock.ExpectQuery(`INSERT INTO "content"."content_versions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	mock.ExpectExec(`UPDATE "content"."contents" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// get() after commit.
	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "Old", "test", nil, &oldBody, nil, 10, "draft", "public", "allowed", nil, 0, 0, json.RawMessage("{}"), 0, now, now, nil))
	mock.ExpectQuery(`SELECT "username" FROM "identity"."users" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("admin"))
	mock.ExpectQuery(`SELECT \* FROM "content"."content_tags" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"content_id", "tag_id", "created_at"}))
	mock.ExpectQuery(`SELECT \* FROM "content"."content_categories" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"content_id", "category_id", "created_at"}))

	ctx, rec := testCrudContextWithID(http.MethodPost, "/api/v1/contents/1/versions/1/restore", bytes.NewReader(body), "1")
	ctx.Params = append(ctx.Params, gin.Param{Key: "versionNumber", Value: "1"})
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 10, Role: "admin"})

	c.RestoreVersion(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRestoreVersionInvalidVersionNumber(t *testing.T) {
	c, _ := newCrudTestController(t)
	body, _ := json.Marshal(v1.RestoreVersionRequest{})

	ctx, rec := testCrudContextWithID(http.MethodPost, "/api/v1/contents/1/versions/abc/restore", bytes.NewReader(body), "1")
	ctx.Params = append(ctx.Params, gin.Param{Key: "versionNumber", Value: "abc"})
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 10, Role: "admin"})

	c.RestoreVersion(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusBadRequest, "versionNumber must be a positive integer")
}
