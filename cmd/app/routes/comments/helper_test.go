package comments

import "testing"

func TestNormalizeEmailHashIsCaseInsensitive(t *testing.T) {
	first := normalizeEmailHash("Reader@Example.COM")
	second := normalizeEmailHash(" reader@example.com ")
	if first != second {
		t.Fatalf("hash mismatch: %q != %q", first, second)
	}
}

func TestCleanOptionalWebsiteRejectsUnsafeScheme(t *testing.T) {
	raw := "javascript:alert(1)"
	if _, err := cleanOptionalWebsite(&raw); err == nil {
		t.Fatal("expected unsafe website URL to be rejected")
	}
}
