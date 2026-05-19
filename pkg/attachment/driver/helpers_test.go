package driver

import (
	"strings"
	"testing"
	"time"
)

func TestCleanObjectKey(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "normal", in: "content/2026/01/02/abc.png", want: "content/2026/01/02/abc.png"},
		{name: "backslash", in: `content\file.png`, want: "content/file.png"},
		{name: "dotdot_segment", in: "content/../secret", wantErr: true},
		{name: "leading_slash", in: "/content/a.png", wantErr: true},
		{name: "empty", in: "  ", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cleanObjectKey(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatal("cleanObjectKey: expected error")
				}
				if err != ErrUnsafeObjectKey {
					t.Fatalf("cleanObjectKey error = %v, want ErrUnsafeObjectKey", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("cleanObjectKey: %v", err)
			}
			if got != tt.want {
				t.Fatalf("cleanObjectKey = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPathEscapeObjectKey(t *testing.T) {
	got := pathEscapeObjectKey("a b/c d")
	if got != "a%20b/c%20d" {
		t.Fatalf("pathEscapeObjectKey = %q, want a%%20b/c%%20d", got)
	}
}

func TestJoinObjectURL(t *testing.T) {
	expires := time.Unix(1700000000, 0)
	url := joinObjectURL("https://cdn.example.com/uploads", "content/2026/file.png", expires)
	if !strings.HasPrefix(url, "https://cdn.example.com/uploads/content/2026/file.png") {
		t.Fatalf("unexpected url prefix: %q", url)
	}
	if !strings.Contains(url, "expires=1700000000") {
		t.Fatalf("url missing expires query: %q", url)
	}
}
