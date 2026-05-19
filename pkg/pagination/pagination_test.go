package pagination

import "testing"

func TestNormalizeOffsetDefaults(t *testing.T) {
	p, ps := NormalizeOffset(0, 0)
	if p != 1 {
		t.Fatalf("page = %d, want 1", p)
	}
	if ps != DefaultPageSize {
		t.Fatalf("pageSize = %d, want %d", ps, DefaultPageSize)
	}
}

func TestNormalizeOffsetKeepsValidValues(t *testing.T) {
	p, ps := NormalizeOffset(3, 50)
	if p != 3 {
		t.Fatalf("page = %d, want 3", p)
	}
	if ps != 50 {
		t.Fatalf("pageSize = %d, want 50", ps)
	}
}

func TestNormalizeOffsetClampsNegative(t *testing.T) {
	p, ps := NormalizeOffset(-5, -1)
	if p != 1 {
		t.Fatalf("page = %d, want 1", p)
	}
	if ps != DefaultPageSize {
		t.Fatalf("pageSize = %d, want %d", ps, DefaultPageSize)
	}
}
