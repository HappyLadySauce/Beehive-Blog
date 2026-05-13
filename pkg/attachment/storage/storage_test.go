package storage_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/attachment/storage"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

func TestLocalBackendRejectsTraversal(t *testing.T) {
	backend, err := storage.NewLocalBackend(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocalBackend: %v", err)
	}
	if _, err := backend.LocalFilePath("../secret.txt"); !errors.Is(err, storage.ErrUnsafeObjectKey) {
		t.Fatalf("LocalFilePath traversal error = %v, want ErrUnsafeObjectKey", err)
	}
}

func TestLocalBackendSaveWritesChecksum(t *testing.T) {
	backend, err := storage.NewLocalBackend(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocalBackend: %v", err)
	}
	out, err := backend.Save(context.Background(), storage.PutRequest{
		ObjectKey: "content/2026/05/11/file.txt",
		Reader:    strings.NewReader("hello"),
		Size:      5,
	})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if out.LocalPath == "" || out.ETag == "" || !strings.HasPrefix(out.Checksum, "sha256:") {
		t.Fatalf("unexpected stored object: %+v", out)
	}
}

func TestRemoteBackendPresignRequiresConfiguration(t *testing.T) {
	backend := storage.NewRemoteBackend(options.AttachmentStorageS3, options.AttachmentRemoteOptions{})
	if _, err := backend.PresignUpload(context.Background(), storage.PresignRequest{
		ObjectKey: "content/file.png",
		MimeType:  "image/png",
		TTL:       time.Minute,
	}); err == nil {
		t.Fatal("expected error for unconfigured remote backend")
	}
}
