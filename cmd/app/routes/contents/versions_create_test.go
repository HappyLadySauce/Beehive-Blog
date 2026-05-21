package contents

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
)

func TestCreateVersionRequiresManualName(t *testing.T) {
	c, _ := newCrudTestController(t)
	body, _ := json.Marshal(v1.CreateVersionRequest{})

	ctx, rec := testCrudContextWithID(http.MethodPost, "/api/v1/contents/1/versions", bytes.NewReader(body), "1")
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 10, Role: "admin"})

	c.CreateVersion(ctx)
	env := decodeCrudEnvelope(t, rec)
	assertCrudError(t, rec, env, http.StatusBadRequest, "version name is required")
}

func TestCreateVersionManualCreatesNamedSnapshot(t *testing.T) {
	c, mock := newCrudTestController(t)
	name := "发布前版本"
	summary := "ready for review"
	body, _ := json.Marshal(v1.CreateVersionRequest{Name: &name, ChangeSummary: &summary})

	now := time.Now()
	contentBody := "Current body"
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "Current Title", "test", nil, &contentBody, nil, 10, "draft", "public", "allowed", nil, 0, 0, json.RawMessage("{}"), 0, now, now, nil))
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(version_number\), 0\) FROM "content"."content_versions" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(1))
	mock.ExpectQuery(`INSERT INTO "content"."content_versions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectCommit()

	ctx, rec := testCrudContextWithID(http.MethodPost, "/api/v1/contents/1/versions", bytes.NewReader(body), "1")
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 10, Role: "admin"})

	c.CreateVersion(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.CreateVersionResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if resp.Name != name {
		t.Fatalf("name = %q, want %q", resp.Name, name)
	}
	if resp.SnapshotType != "manual" {
		t.Fatalf("snapshot_type = %q, want manual", resp.SnapshotType)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateVersionAutoCreatesFirstSnapshot(t *testing.T) {
	c, mock := newCrudTestController(t)
	snapshotType := "auto"
	body, _ := json.Marshal(v1.CreateVersionRequest{SnapshotType: &snapshotType})

	now := time.Now()
	contentBody := "Current body"
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "Current Title", "test", nil, &contentBody, nil, 10, "draft", "public", "allowed", nil, 0, 0, json.RawMessage("{}"), 0, now, now, nil))
	mock.ExpectQuery(`SELECT \* FROM "content"."content_versions" WHERE`).
		WillReturnRows(sqlmock.NewRows(versionColumns()))
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(version_number\), 0\) FROM "content"."content_versions" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(0))
	mock.ExpectQuery(`INSERT INTO "content"."content_versions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	ctx, rec := testCrudContextWithID(http.MethodPost, "/api/v1/contents/1/versions", bytes.NewReader(body), "1")
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 10, Role: "admin"})

	c.CreateVersion(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.CreateVersionResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if resp.Name != autoVersionName || resp.SnapshotType != "auto" {
		t.Fatalf("auto version = (%q, %q), want (%q, auto)", resp.Name, resp.SnapshotType, autoVersionName)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateVersionAutoUpdatesExistingSnapshot(t *testing.T) {
	c, mock := newCrudTestController(t)
	snapshotType := "auto"
	body, _ := json.Marshal(v1.CreateVersionRequest{SnapshotType: &snapshotType})

	now := time.Now()
	contentBody := "Newest body"
	oldBody := "Older body"
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "content"."contents" WHERE`).
		WillReturnRows(sqlmock.NewRows(contentColumns()).
			AddRow(1, "article", "Newest Title", "test", nil, &contentBody, nil, 10, "draft", "public", "allowed", nil, 0, 0, json.RawMessage("{}"), 0, now, now, nil))
	mock.ExpectQuery(`SELECT \* FROM "content"."content_versions" WHERE`).
		WillReturnRows(sqlmock.NewRows(versionColumns()).
			AddRow(4, 1, 1, "auto", autoVersionName, "Older Title", &oldBody, nil, nil, 10, now))
	mock.ExpectExec(`UPDATE "content"."content_versions" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx, rec := testCrudContextWithID(http.MethodPost, "/api/v1/contents/1/versions", bytes.NewReader(body), "1")
	ctx.Set(jwt.ClaimsKey, &jwt.Claims{UID: 10, Role: "admin"})

	c.CreateVersion(ctx)
	env := decodeCrudEnvelope(t, rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP = %d, want 200", rec.Code)
	}
	var resp v1.CreateVersionResponse
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if resp.ID != 4 {
		t.Fatalf("id = %d, want existing auto version id 4", resp.ID)
	}
	if resp.Title != "Newest Title" || resp.Body == nil || *resp.Body != contentBody {
		t.Fatalf("auto version did not return latest content snapshot")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
