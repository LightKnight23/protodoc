package integrity

import (
	"errors"
	"testing"
)

// TestFR_063_PDCover001RejectsZeroLengthRange is T-0146's named conformance
// test for rule PD-COVER-001: a coverage range whose end does not exceed its
// start (a zero-length or inverted range) is rejected, not silently dropped,
// in either the covered or the uncovered list.
func TestFR_063_PDCover001RejectsZeroLengthRange(t *testing.T) {
	noAttest := map[uint16]bool{}

	cases := []struct {
		name string
		d    CoverageDescriptor
		segs int
	}{
		{
			name: "zero-length in covered [3,3)",
			d:    CoverageDescriptor{Mode: CoverageModeSubset, Covered: []SegmentRange{{Start: 3, End: 3}}},
			segs: 10,
		},
		{
			name: "inverted in covered [5,2)",
			d:    CoverageDescriptor{Mode: CoverageModeSubset, Covered: []SegmentRange{{Start: 5, End: 2}}},
			segs: 10,
		},
		{
			name: "zero-length in uncovered [0,0)",
			d:    CoverageDescriptor{Mode: CoverageModeSubset, Uncovered: []SegmentRange{{Start: 0, End: 0}}},
			segs: 10,
		},
	}
	for _, c := range cases {
		err := c.d.Validate(c.segs, noAttest)
		if !errors.Is(err, ErrCoverZeroLengthRange) {
			t.Errorf("%s: Validate err = %v, want ErrCoverZeroLengthRange (PD-COVER-001)", c.name, err)
		}
	}

	// A well-formed positive-length range at the same site is accepted (the
	// rule rejects only the degenerate ranges, not all ranges).
	ok := CoverageDescriptor{
		Mode:      CoverageModeSubset,
		Covered:   []SegmentRange{{Start: 0, End: 3}},
		Uncovered: []SegmentRange{{Start: 3, End: 10}},
	}
	if err := ok.Validate(10, noAttest); err != nil {
		t.Errorf("well-formed descriptor rejected: %v", err)
	}
}
