package attachment

import (
	"errors"
	"testing"

	"gorm.io/gorm"
)

func TestMapDBErrorNil(t *testing.T) {
	if err := MapDBError(nil); err != nil {
		t.Fatalf("MapDBError(nil) = %v, want nil", err)
	}
}

func TestMapDBErrorMapsRecordNotFound(t *testing.T) {
	err := MapDBError(gorm.ErrRecordNotFound)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("MapDBError(ErrRecordNotFound) = %v, want ErrNotFound", err)
	}
}

func TestMapDBErrorPassesThroughOtherErrors(t *testing.T) {
	orig := errors.New("db timeout")
	if err := MapDBError(orig); !errors.Is(err, orig) {
		t.Fatalf("MapDBError() = %v, want %v", err, orig)
	}
}

func TestSafeFilename(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "strips_path", in: `foo\bar\file.txt`, want: "file.txt"},
		{name: "replaces_spaces", in: "hello world.png", want: "hello-world.png"},
		{name: "empty", in: "   ", want: ""},
		{name: "keeps_safe_chars", in: "a-z_0.9.txt", want: "a-z_0.9.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeFilename(tt.in); got != tt.want {
				t.Fatalf("SafeFilename(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestObjectKeyForRejectsEmptyFilename(t *testing.T) {
	_, _, err := ObjectKeyFor(PurposeContent, "  ")
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("ObjectKeyFor error = %v, want ErrInvalid", err)
	}
}

func TestCleanOptional(t *testing.T) {
	if CleanOptional(nil) != nil {
		t.Fatal("CleanOptional(nil) should be nil")
	}
	blank := "   "
	if CleanOptional(&blank) != nil {
		t.Fatal("CleanOptional(blank) should be nil")
	}
	s := "  x  "
	got := CleanOptional(&s)
	if got == nil || *got != "x" {
		t.Fatalf("CleanOptional(trim) = %v, want x", got)
	}
}

func TestDerefString(t *testing.T) {
	if DerefString(nil) != "" {
		t.Fatal("DerefString(nil) should be empty")
	}
	s := "  hi "
	if got := DerefString(&s); got != "hi" {
		t.Fatalf("DerefString = %q, want hi", got)
	}
}

func TestUniqueInt64PreservesOrderAndDedupes(t *testing.T) {
	got := UniqueInt64([]int64{3, 1, 3, 2, 1})
	want := []int64{3, 1, 2}
	if len(got) != len(want) {
		t.Fatalf("UniqueInt64 = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("UniqueInt64 = %v, want %v", got, want)
		}
	}
}

func TestUniqueInt64Empty(t *testing.T) {
	if got := UniqueInt64(nil); len(got) != 0 {
		t.Fatalf("UniqueInt64(nil) = %v, want empty slice", got)
	}
}
