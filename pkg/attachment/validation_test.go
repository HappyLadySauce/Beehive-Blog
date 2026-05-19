package attachment

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateCommonRejectsAvatarMime(t *testing.T) {
	err := ValidateCommon(int64Ptr(10), PurposeAvatar, "text/plain", 5, AccessPrivate)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("ValidateCommon error = %v, want ErrInvalid", err)
	}
}

func TestValidateCommonAllowsAnyNonAvatarMime(t *testing.T) {
	err := ValidateCommon(int64Ptr(10), PurposeContent, "application/x-custom", 5, AccessPrivate)
	if err != nil {
		t.Fatalf("ValidateCommon error = %v, want nil", err)
	}
}

func TestValidateCommonRejectsFilesOverTwoGiB(t *testing.T) {
	err := ValidateCommon(int64Ptr(10), PurposeContent, "application/octet-stream", MaxUploadBytes+1, AccessPrivate)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("ValidateCommon error = %v, want ErrInvalid", err)
	}
}

func TestValidateCommon(t *testing.T) {
	tests := []struct {
		name        string
		ownerUserID *int64
		purpose     string
		mimeType    string
		size        int64
		accessScope string
		wantErr     bool
	}{
		{
			name:        "system_without_owner",
			ownerUserID: nil,
			purpose:     PurposeSystem,
			mimeType:    "application/octet-stream",
			size:        1,
			accessScope: AccessPublic,
		},
		{
			name:        "non_system_requires_owner",
			ownerUserID: nil,
			purpose:     PurposeContent,
			mimeType:    "image/png",
			size:        1,
			accessScope: AccessPrivate,
			wantErr:     true,
		},
		{
			name:        "invalid_purpose",
			ownerUserID: int64Ptr(1),
			purpose:     "bogus",
			mimeType:    "image/png",
			size:        1,
			accessScope: AccessPrivate,
			wantErr:     true,
		},
		{
			name:        "empty_mime",
			ownerUserID: int64Ptr(1),
			purpose:     PurposeContent,
			mimeType:    "  ",
			size:        1,
			accessScope: AccessPrivate,
			wantErr:     true,
		},
		{
			name:        "zero_size",
			ownerUserID: int64Ptr(1),
			purpose:     PurposeContent,
			mimeType:    "image/png",
			size:        0,
			accessScope: AccessPrivate,
			wantErr:     true,
		},
		{
			name:        "invalid_access_scope",
			ownerUserID: int64Ptr(1),
			purpose:     PurposeContent,
			mimeType:    "image/png",
			size:        1,
			accessScope: "shared",
			wantErr:     true,
		},
		{
			name:        "avatar_image_ok",
			ownerUserID: int64Ptr(1),
			purpose:     PurposeAvatar,
			mimeType:    "image/jpeg",
			size:        1024,
			accessScope: AccessPrivate,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommon(tt.ownerUserID, tt.purpose, tt.mimeType, tt.size, tt.accessScope)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("ValidateCommon error = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateCommon error = %v, want nil", err)
			}
		})
	}
}

func TestObjectKeyForSanitizesFilename(t *testing.T) {
	objectKey, safeName, err := ObjectKeyFor(PurposeContent, `..\hello world.png`)
	if err != nil {
		t.Fatalf("ObjectKeyFor: %v", err)
	}
	if safeName != "hello-world.png" {
		t.Fatalf("safeName = %q, want hello-world.png", safeName)
	}
	if !strings.HasPrefix(objectKey, PurposeContent+"/") || !strings.HasSuffix(objectKey, ".png") {
		t.Fatalf("unexpected objectKey: %q", objectKey)
	}
}

func TestRequireAdmin(t *testing.T) {
	if err := RequireAdmin(Actor{UID: 10, Role: RoleAdmin}); err != nil {
		t.Fatalf("RequireAdmin(admin) = %v, want nil", err)
	}
	if err := RequireAdmin(Actor{UID: 10, Role: "member"}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("RequireAdmin(member) error = %v, want ErrForbidden", err)
	}
}

func TestPurposeKnown(t *testing.T) {
	for _, p := range []string{PurposeAvatar, PurposeContent, PurposeSystem, PurposeOther} {
		if !PurposeKnown(p) {
			t.Fatalf("PurposeKnown(%q) = false, want true", p)
		}
	}
	if PurposeKnown("unknown") {
		t.Fatal("PurposeKnown(unknown) = true, want false")
	}
}

func TestAccessScopeKnown(t *testing.T) {
	if !AccessScopeKnown(AccessPrivate) || !AccessScopeKnown(AccessPublic) {
		t.Fatal("expected private and public scopes to be known")
	}
	if AccessScopeKnown("link") {
		t.Fatal("AccessScopeKnown(link) = true, want false")
	}
}

func TestStatusKnown(t *testing.T) {
	for _, s := range []string{StatusActive, StatusHidden, StatusArchived} {
		if !StatusKnown(s) {
			t.Fatalf("StatusKnown(%q) = false, want true", s)
		}
	}
	if StatusKnown("deleted") {
		t.Fatal("StatusKnown(deleted) = true, want false")
	}
}

func TestCategoryStatusKnown(t *testing.T) {
	if !CategoryStatusKnown(CategoryStatusActive) || !CategoryStatusKnown(CategoryStatusDisabled) {
		t.Fatal("expected category statuses to be known")
	}
	if CategoryStatusKnown("pending") {
		t.Fatal("CategoryStatusKnown(pending) = true, want false")
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}
