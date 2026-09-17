package integrity

import (
	"errors"
	"testing"
)

// TestFR_063_PDCover004RejectsAttestOrdinal is T-0149's named conformance test
// for rule PD-COVER-004: an ATTEST-typed ordinal is never nameable in a
// coverage list, in either list, for any signature (closing FR-002's
// self-coverage circularity). Naming one is rejected.
func TestFR_063_PDCover004RejectsAttestOrdinal(t *testing.T) {
	const segs = 10
	// Ordinal 7 is ATTEST-typed.
	attest := map[uint16]bool{7: true}

	cases := []struct {
		name string
		d    CoverageDescriptor
	}{
		{
			name: "ATTEST ordinal 7 named in covered",
			d: CoverageDescriptor{
				Mode:    CoverageModeSubset,
				Covered: []SegmentRange{{Start: 6, End: 9}}, // includes 7 (ATTEST)
			},
		},
		{
			name: "ATTEST ordinal 7 named in uncovered",
			d: CoverageDescriptor{
				Mode:      CoverageModeSubset,
				Uncovered: []SegmentRange{{Start: 7, End: 8}}, // exactly the ATTEST ordinal
			},
		},
	}
	for _, c := range cases {
		err := c.d.Validate(segs, attest)
		if !errors.Is(err, ErrCoverAttestNamed) {
			t.Errorf("%s: Validate err = %v, want ErrCoverAttestNamed (PD-COVER-004)", c.name, err)
		}
	}

	// A descriptor that partitions every NON-ATTEST ordinal and never names
	// the ATTEST ordinal 7 is accepted (7 is excluded from the partition).
	ok := CoverageDescriptor{
		Mode:      CoverageModeSubset,
		Covered:   []SegmentRange{{Start: 0, End: 4}},
		Uncovered: []SegmentRange{{Start: 4, End: 7}, {Start: 8, End: segs}},
	}
	if err := ok.Validate(segs, attest); err != nil {
		t.Errorf("descriptor correctly excluding the ATTEST ordinal rejected: %v", err)
	}
}
