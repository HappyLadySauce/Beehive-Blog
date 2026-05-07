package options

import (
	"strings"
	"testing"
)

func TestInsecureServingValidateJoinsAllMissingFields(t *testing.T) {
	t.Helper()

	opts := &InsecureServingOptions{}
	err := opts.Validate()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "bind-address is required") {
		t.Fatalf("expected bind-address validation error, got %v", err)
	}
	if !strings.Contains(err.Error(), "bind-port is required") {
		t.Fatalf("expected bind-port validation error, got %v", err)
	}

	joined, ok := err.(interface{ Unwrap() []error })
	if !ok {
		t.Fatalf("expected joined error, got %T", err)
	}
	if got := len(joined.Unwrap()); got != 2 {
		t.Fatalf("expected 2 joined errors, got %d", got)
	}
}

func TestInsecureServingValidateReturnsSingleErrorWhenOneFieldMissing(t *testing.T) {
	t.Helper()

	opts := &InsecureServingOptions{BindAddress: "127.0.0.1"}
	err := opts.Validate()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if got := err.Error(); got != "bind-port is required" {
		t.Fatalf("expected single bind-port validation error, got %q", got)
	}
}
