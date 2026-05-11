package attachment

import (
	"errors"
	"strings"
	"testing"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

func TestValidateCommonRejectsAvatarMime(t *testing.T) {
	opts := options.NewAttachmentOptions()
	err := ValidateCommon(opts, int64Ptr(10), PurposeAvatar, "text/plain", 5, AccessPrivate)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("ValidateCommon error = %v, want ErrInvalid", err)
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

func TestRequireAdminRejectsNonAdmin(t *testing.T) {
	if err := RequireAdmin(Actor{UID: 10, Role: "member"}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("RequireAdmin error = %v, want ErrForbidden", err)
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}
