package content

import (
	"errors"
	"testing"
)

// TestFR_026_BoundaryBehaviourIsClosedFourValueEnum is T-0076's named
// conformance test. It exercises the closed 4-value anchor-boundary enum
// (document.abnf S3.1): the four defined values decode to their behaviours,
// and the CON-010/CP-011 at-limit (value 3, last valid) / one-past-limit
// (value 4, first reserved) fixture pair confirms value 3 is accepted and
// value 4 is a structural rejection (rule PD-BOUNDARY-001).
func TestFR_026_BoundaryBehaviourIsClosedFourValueEnum(t *testing.T) {
	// The four defined values decode successfully and round-trip.
	valid := []struct {
		octet byte
		want  AnchorBoundary
		name  string
	}{
		{0x00, BoundaryInside, "inside"},
		{0x01, BoundaryOutside, "outside"},
		{0x02, BoundaryInsideIfInsertedBefore, "inside-if-inserted-before"},
		{0x03, BoundaryInsideIfInsertedAfter, "inside-if-inserted-after"},
	}
	for _, c := range valid {
		got, err := DecodeAnchorBoundary(c.octet)
		if err != nil {
			t.Fatalf("DecodeAnchorBoundary(0x%02x): unexpected error %v", c.octet, err)
		}
		if got != c.want {
			t.Fatalf("DecodeAnchorBoundary(0x%02x) = %v, want %v", c.octet, got, c.want)
		}
		if got.String() != c.name {
			t.Fatalf("boundary 0x%02x String() = %q, want %q", c.octet, got.String(), c.name)
		}
	}

	// At-limit fixture: value 3 (the last valid value) is accepted.
	if _, err := DecodeAnchorBoundary(0x03); err != nil {
		t.Fatalf("at-limit fixture (value 3) was rejected: %v", err)
	}

	// One-past-limit fixture: value 4 (first reserved value) is a structural
	// rejection naming PD-BOUNDARY-001.
	if _, err := DecodeAnchorBoundary(0x04); !errors.Is(err, ErrInvalidAnchorBoundary) {
		t.Fatalf("over-limit fixture (value 4) returned %v, want ErrInvalidAnchorBoundary", err)
	}

	// Every reserved value 0x04..0xFF is rejected (the set is closed).
	for v := 0x04; v <= 0xFF; v++ {
		if _, err := DecodeAnchorBoundary(byte(v)); !errors.Is(err, ErrInvalidAnchorBoundary) {
			t.Fatalf("reserved value 0x%02x was not rejected", v)
		}
	}
}
