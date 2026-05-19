package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestAttachmentSchemaMigrationsFinalizeStorageMountLocator(t *testing.T) {
	t.Helper()

	root := repoRootFromWorkingDir(t)
	finalSQL := readRepoFile(t, root, filepath.Join("sql", "migrations", "attachment", "007_attachment_finalize_storage_mounts.sql"))

	if !strings.Contains(finalSQL, "ALTER COLUMN storage_mount_id SET NOT NULL") {
		t.Fatalf("final storage migration should require storage_mount_id")
	}
	if !strings.Contains(finalSQL, "ALTER COLUMN object_key SET NOT NULL") {
		t.Fatalf("final storage migration should require object_key")
	}
	for _, oldColumn := range []string{"DROP COLUMN storage_type", "DROP COLUMN bucket", "DROP COLUMN local_path"} {
		if !strings.Contains(finalSQL, oldColumn) {
			t.Fatalf("final storage migration should remove old column statement %q", oldColumn)
		}
	}
	if !strings.Contains(finalSQL, "CREATE UNIQUE INDEX ux_attachment_attachments_mount_object_key") {
		t.Fatalf("final storage migration should define mount + object_key uniqueness")
	}
	if !strings.Contains(finalSQL, "ON attachment.attachments (storage_mount_id, created_at DESC, id DESC)") {
		t.Fatalf("final storage migration should list by storage_mount_id, created_at DESC, id DESC")
	}
	for _, legacyIndex := range []string{
		"idx_attachment_attachments_live_storage_type_created_at",
		"ux_attachment_attachments_remote_object",
		"ux_attachment_attachments_local_path",
	} {
		if !strings.Contains(finalSQL, legacyIndex) {
			t.Fatalf("final storage migration should explicitly drop legacy index %q", legacyIndex)
		}
	}
}

func TestAttachmentSchemaMigrationsKeepPurposeAndAddOwnerFKAfterIdentity(t *testing.T) {
	t.Helper()

	root := repoRootFromWorkingDir(t)
	attachmentSQL := readRepoFile(t, root, filepath.Join("sql", "migrations", "attachment", "000_attachment_attachments.sql"))
	ownerFKSQL := readRepoFile(t, root, filepath.Join("sql", "migrations", "attachment", "014_attachment_attachments_owner_fk.sql"))

	for _, want := range []string{
		"purpose         VARCHAR(32) NOT NULL DEFAULT 'content'",
		"CONSTRAINT chk_attachment_purpose",
		"CHECK (purpose IN ('avatar', 'content', 'system', 'other'))",
		"CHECK (owner_user_id IS NOT NULL OR purpose = 'system')",
		"CHECK (purpose <> 'avatar' OR mime_type LIKE 'image/%')",
	} {
		if !strings.Contains(attachmentSQL, want) {
			t.Fatalf("attachment migration should keep purpose semantics, missing %q", want)
		}
	}
	if strings.Contains(attachmentSQL, "usage_type") {
		t.Fatalf("attachment migration should use purpose, not usage_type")
	}
	for _, want := range []string{
		"ADD CONSTRAINT fk_attachment_attachments_owner_user",
		"FOREIGN KEY (owner_user_id)",
		"REFERENCES identity.users (id)",
		"ON DELETE RESTRICT",
	} {
		if !strings.Contains(ownerFKSQL, want) {
			t.Fatalf("owner FK migration missing %q", want)
		}
	}
}

func TestIdentitySchemaMigrationsAddRefreshSessionsAndProviderIdentities(t *testing.T) {
	t.Helper()

	root := repoRootFromWorkingDir(t)
	sessionSQL := readRepoFile(t, root, filepath.Join("sql", "migrations", "identity", "012_identity_user_sessions.sql"))
	identitySQL := readRepoFile(t, root, filepath.Join("sql", "migrations", "identity", "013_identity_user_identities.sql"))

	for _, want := range []string{
		"CREATE TABLE identity.user_sessions",
		"refresh_token_hash CHAR(64) NOT NULL",
		"refresh_jti VARCHAR(64) NOT NULL",
		"rotated_at TIMESTAMPTZ NULL",
		"ux_user_sessions_refresh_jti",
	} {
		if !strings.Contains(sessionSQL, want) {
			t.Fatalf("user sessions migration missing %q", want)
		}
	}
	for _, want := range []string{
		"CREATE TABLE identity.user_identities",
		"provider VARCHAR(32) NOT NULL",
		"provider_subject VARCHAR(128) NOT NULL",
		"ux_user_identities_provider_subject",
	} {
		if !strings.Contains(identitySQL, want) {
			t.Fatalf("user identities migration missing %q", want)
		}
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
