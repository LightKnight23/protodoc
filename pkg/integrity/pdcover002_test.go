package integrity

import (
	"errors"
	"testing"
)

// TestFR_063_PDCover002RejectsUnmergedAdjacentRanges is T-0147's named
// conformance test for rule PD-COVER-002: merge-adjacent canonicalisation is
// mandatory. Two entries (a,b) and (b,c) in the SAME list must be a single
// (a,c); a technically-sorted-but-unmerged pair, or an unsorted/overlapping
// pair, is rejected -- exactly one valid encoding exists per coverage set.
func TestFR_063_PDCover002RejectsUnmergedAdjacentRanges(t *testing.T) {
	noAttest := map[uint16]bool{}

	cases := []struct {
		name string
		d    CoverageDescriptor
	}{
		{
			name: "adjacent unmerged covered (0,2)(2,4) should be (0,4)",
			d:    CoverageDescriptor{Mode: CoverageModeSubset, Covered: []SegmentRange{{Start: 0, End: 2}, {Start: 2, End: 4}}},
		},
		{
			name: "overlapping covered (0,3)(2,5)",
			d:    CoverageDescriptor{Mode: CoverageModeSubset, Covered: []SegmentRange{{Start: 0, End: 3}, {Start: 2, End: 5}}},
		},
		{
			name: "unsorted covered (5,7)(0,2)",
			d:    CoverageDescriptor{Mode: CoverageModeSubset, Covered: []SegmentRange{{Start: 5, End: 7}, {Start: 0, End: 2}}},
		},
		{
			name: "adjacent unmerged uncovered (3,5)(5,9)",
			d:    CoverageDescriptor{Mode: CoverageModeSubset, Uncovered: []SegmentRange{{Start: 3, End: 5}, {Start: 5, End: 9}}},
		},
	}
	for _, c := range cases {
		err := c.d.Validate(20, noAttest)
		if !errors.Is(err, ErrCoverNotCanonical) {
			t.Errorf("%s: Validate err = %v, want ErrCoverNotCanonical (PD-COVER-002)", c.name, err)
		}
	}

	// The merged canonical form of an adjacent pair is accepted.
	ok := CoverageDescriptor{
		Mode:      CoverageModeSubset,
		Covered:   []SegmentRange{{Start: 0, End: 4}}, // the merge of (0,2)(2,4)
		Uncovered: []SegmentRange{{Start: 4, End: 20}},
	}
	if err := ok.Validate(20, noAttest); err != nil {
		t.Errorf("canonical merged descriptor rejected: %v", err)
	}
}
