package integrity

import (
	"errors"
	"testing"
)

// TestFR_063_PDCover003RejectsGapOrOverlap is T-0148's named conformance test
// for rule PD-COVER-003: over [0, segment_count) no ordinal may appear in both
// lists (overlap) and no ordinal may appear in neither (gap); a gap or overlap
// is rejected before any signature verification step.
func TestFR_063_PDCover003RejectsGapOrOverlap(t *testing.T) {
	noAttest := map[uint16]bool{}
	const segs = 10

	cases := []struct {
		name string
		d    CoverageDescriptor
	}{
		{
			name: "gap: ordinal 5..9 covered by neither list",
			d: CoverageDescriptor{
				Mode:      CoverageModeSubset,
				Covered:   []SegmentRange{{Start: 0, End: 3}},
				Uncovered: []SegmentRange{{Start: 3, End: 5}},
				// ordinals 5..9 uncovered by neither list -> gap
			},
		},
		{
			name: "overlap: ordinal 2 in both lists",
			d: CoverageDescriptor{
				Mode:      CoverageModeSubset,
				Covered:   []SegmentRange{{Start: 0, End: 6}},
				Uncovered: []SegmentRange{{Start: 2, End: 10}},
			},
		},
	}
	for _, c := range cases {
		err := c.d.Validate(segs, noAttest)
		if !errors.Is(err, ErrCoverPartition) {
			t.Errorf("%s: Validate err = %v, want ErrCoverPartition (PD-COVER-003)", c.name, err)
		}
	}

	// A complete, non-overlapping partition of [0,segs) is accepted.
	ok := CoverageDescriptor{
		Mode:      CoverageModeSubset,
		Covered:   []SegmentRange{{Start: 0, End: 4}},
		Uncovered: []SegmentRange{{Start: 4, End: segs}},
	}
	if err := ok.Validate(segs, noAttest); err != nil {
		t.Errorf("complete partition rejected: %v", err)
	}
}
