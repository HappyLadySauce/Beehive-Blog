package options_test

import (
	"testing"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

func TestAttachmentOptionsValidateDefaults(t *testing.T) {
	o := options.NewAttachmentOptions()
	if err := o.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestAttachmentOptionsValidateRejectsInvalidStorageAndSize(t *testing.T) {
	o := options.NewAttachmentOptions()
	o.DefaultStorage = "ftp"
	o.MaxBytes = 0
	if err := o.Validate(); err == nil {
		t.Fatal("expected error for invalid storage and size")
	}
}

func TestAttachmentOptionsValidateRemoteRequiresCompleteURLs(t *testing.T) {
	o := options.NewAttachmentOptions()
	o.S3.Bucket = "bucket"
	o.S3.UploadBaseURL = "https://upload.example.com"
	if err := o.Validate(); err == nil {
		t.Fatal("expected error for incomplete s3 remote options")
	}

	o.S3.DownloadBaseURL = "https://download.example.com"
	o.PresignTTL = time.Minute
	if err := o.Validate(); err != nil {
		t.Fatalf("Validate with complete remote options: %v", err)
	}
}
