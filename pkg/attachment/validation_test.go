package attachment_test

import (
	"errors"
	"strings"
	"testing"

	pkgattachment "github.com/HappyLadySauce/Beehive-Blog/pkg/attachment"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

func TestValidateCommonRejectsAvatarMime(t *testing.T) {
	opts := options.NewAttachmentOptions()
	err := pkgattachment.ValidateCommon(opts, int64Ptr(10), pkgattachment.PurposeAvatar, "text/plain", 5, pkgattachment.AccessPrivate)
	if !errors.Is(err, pkgattachment.ErrInvalid) {
		t.Fatalf("ValidateCommon error = %v, want ErrInvalid", err)
	}
}

func TestObjectKeyForSanitizesFilename(t *testing.T) {
	objectKey, safeName, err := pkgattachment.ObjectKeyFor(pkgattachment.PurposeContent, `..\hello world.png`)
	if err != nil {
		t.Fatalf("ObjectKeyFor: %v", err)
	}
	if safeName != "hello-world.png" {
		t.Fatalf("safeName = %q, want hello-world.png", safeName)
	}
	if !strings.HasPrefix(objectKey, pkgattachment.PurposeContent+"/") || !strings.HasSuffix(objectKey, ".png") {
		t.Fatalf("unexpected objectKey: %q", objectKey)
	}
}

func TestRequireAdminRejectsNonAdmin(t *testing.T) {
	if err := pkgattachment.RequireAdmin(pkgattachment.Actor{UID: 10, Role: "member"}); !errors.Is(err, pkgattachment.ErrForbidden) {
		t.Fatalf("RequireAdmin error = %v, want ErrForbidden", err)
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}
