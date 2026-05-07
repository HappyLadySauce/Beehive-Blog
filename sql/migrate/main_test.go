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
	identitySQL := readIdentityUsersMigration(t, root)
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
	identitySQL := readIdentityUsersMigration(t, root)
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

func TestAttachmentSchemaMigrationsOptimizeForRemoteTupleLookupAndStorageTypeListing(t *testing.T) {
	t.Helper()

	root := repoRootFromWorkingDir(t)
	attachmentSQL := readRepoFile(t, root, filepath.Join("sql", "migrations", "attachment", "000_attachment_attachments.sql"))

	if !strings.Contains(attachmentSQL, "CREATE INDEX idx_attachment_attachments_live_storage_type_created_at") {
		t.Fatalf("attachment migration should define a live storage_type + created_at listing index")
	}
	if !strings.Contains(attachmentSQL, "ON attachment.attachments (storage_type, created_at DESC, id DESC)") {
		t.Fatalf("attachment migration should order the listing index by storage_type, created_at DESC, id DESC")
	}
	if !strings.Contains(attachmentSQL, "storage_type + bucket + object_key") {
		t.Fatalf("attachment migration should document remote tuple lookup by storage_type + bucket + object_key")
	}
	if strings.Contains(attachmentSQL, "CREATE INDEX idx_attachment_attachments_storage_type") {
		t.Fatalf("attachment migration should remove the live storage_type-only index")
	}
	if strings.Contains(attachmentSQL, "CREATE INDEX idx_attachment_attachments_object_key") {
		t.Fatalf("attachment migration should remove the object_key-only lookup index")
	}
	if strings.Contains(attachmentSQL, "CREATE INDEX idx_attachment_attachments_created_at") {
		t.Fatalf("attachment migration should remove the global live created_at index when no global timeline is required")
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

func readIdentityUsersMigration(t *testing.T, root string) string {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join(root, "sql", "migrations", "identity", "*_identity_users.sql"))
	if err != nil {
		t.Fatalf("glob identity users migrations failed: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected exactly one identity users migration, got %d", len(matches))
	}

	body, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read %s failed: %v", matches[0], err)
	}
	return string(body)
}
