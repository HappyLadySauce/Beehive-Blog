package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListMigrationFilesSortsByFilenameBeforeDirectory(t *testing.T) {
	t.Helper()

	root := t.TempDir()
	migrationsDir := filepath.Join(root, "sql", "migrations", "v3")
	writeMigrationTestFile(t, migrationsDir, filepath.Join("content", "030_v3_content_items.sql"))
	writeMigrationTestFile(t, migrationsDir, filepath.Join("identity", "020_v3_identity_users.sql"))

	files, err := listMigrationFiles(migrationsDir, filepath.Join(root, "sql", "migrations"))
	if err != nil {
		t.Fatalf("list migration files failed: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("unexpected migration count: %d", len(files))
	}
	if files[0].Name != "identity/020_v3_identity_users.sql" {
		t.Fatalf("expected identity migration first, got %s", files[0].Name)
	}
	if files[1].Name != "content/030_v3_content_items.sql" {
		t.Fatalf("expected content migration second, got %s", files[1].Name)
	}
}

func writeMigrationTestFile(t *testing.T, root, relativePath string) {
	t.Helper()

	path := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create migration directory failed: %v", err)
	}
	if err := os.WriteFile(path, []byte("SELECT 1;\n"), 0o644); err != nil {
		t.Fatalf("write migration file failed: %v", err)
	}
}

func TestSchemaMigrationsEncodeAvatarFallbackSemantics(t *testing.T) {
	t.Helper()

	root := repoRootFromWorkingDir(t)
	identitySQL := readRepoFile(t, root, filepath.Join("sql", "migrations", "identity", "001_identity_users.sql"))
	attachmentSQL := readRepoFile(t, root, filepath.Join("sql", "migrations", "attachment", "000_attachment_attachments.sql"))

	if !strings.Contains(identitySQL, "NULL means use the application default avatar") {
		t.Fatalf("identity migration should document that NULL avatar_attachment_id falls back to the application default avatar")
	}
	if !strings.Contains(attachmentSQL, "hidden remains referenceable") {
		t.Fatalf("attachment migration should document that hidden attachments can still be referenced")
	}
}

func TestSchemaMigrationsPlaceAvatarSoftDeleteTriggerWithAttachmentLifecycle(t *testing.T) {
	t.Helper()

	root := repoRootFromWorkingDir(t)
	identitySQL := readRepoFile(t, root, filepath.Join("sql", "migrations", "identity", "001_identity_users.sql"))
	attachmentSQL := readRepoFile(t, root, filepath.Join("sql", "migrations", "attachment", "000_attachment_attachments.sql"))

	if strings.Contains(identitySQL, "trg_attachment_attachments_clear_users_avatar_on_soft_delete") {
		t.Fatalf("identity migration should not define attachment soft-delete trigger")
	}
	if !strings.Contains(attachmentSQL, "trg_attachment_attachments_clear_users_avatar_on_soft_delete") {
		t.Fatalf("attachment migration should define attachment soft-delete trigger")
	}
	if !strings.Contains(attachmentSQL, "updated_at = NOW()") {
		t.Fatalf("attachment migration should refresh user updated_at when avatar falls back to default")
	}
}

func repoRootFromWorkingDir(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory failed: %v", err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

func readRepoFile(t *testing.T, root, relativePath string) string {
	t.Helper()

	body, err := os.ReadFile(filepath.Join(root, relativePath))
	if err != nil {
		t.Fatalf("read %s failed: %v", relativePath, err)
	}
	return string(body)
}
