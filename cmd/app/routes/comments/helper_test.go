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

func TestCleanRequiredTextRejectsWhitespaceOnly(t *testing.T) {
	if _, err := cleanRequiredText(" \n\t ", "body"); err == nil {
		t.Fatal("expected whitespace-only text to be rejected")
	}
}

func TestCleanRequiredTextTrimsValidText(t *testing.T) {
	value, err := cleanRequiredText("  hello  ", "body")
	if err != nil {
		t.Fatalf("expected valid text, got error: %v", err)
	}
	if value != "hello" {
		t.Fatalf("unexpected trimmed value: %q", value)
	}
}
