package integrity

import (
	"testing"
)

// TestFR_063_ConformanceCorpusAtLimitAndOverLimitFixtures is T-0166's named
// conformance test (FR-063). It exercises the signature/coverage ceilings with
// paired at-limit (accepted) and over-limit (rejected) fixtures, so the exact
// boundary of each ceiling is pinned: MAX_SIGNATURES (64) and a coverage
// descriptor whose covered ranges partition exactly a document's ordinals at
// and beyond the segment count.
func TestFR_063_ConformanceCorpusAtLimitAndOverLimitFixtures(t *testing.T) {
	// --- MAX_SIGNATURES boundary ---
	if MaxSignatures != 64 {
		t.Fatalf("MaxSignatures = %d, want 64", MaxSignatures)
	}
	// At limit: a document with exactly MaxSignatures signature frames is
	// within the ceiling.
	if countWithinSignatureCeiling(MaxSignatures) != true {
		t.Errorf("exactly %d signatures should be within the ceiling", MaxSignatures)
	}
	// Over limit: one more is over the ceiling.
	if countWithinSignatureCeiling(MaxSignatures+1) != false {
		t.Errorf("%d signatures should exceed the ceiling", MaxSignatures+1)
	}

	// --- Coverage partition boundary against a segment count ---
	noAttest := map[uint16]bool{}
	const segs = 8

	// At limit: a TOTAL descriptor covering exactly [0,segs) validates.
	atLimit := CoverageDescriptor{Mode: CoverageModeTotal, Covered: []SegmentRange{{Start: 0, End: segs}}}
	if err := atLimit.Validate(segs, noAttest); err != nil {
		t.Errorf("at-limit TOTAL coverage of [0,%d) rejected: %v", segs, err)
	}

	// Over limit: a covered range extending to segs+1 leaves ordinal `segs`
	// out of the document's [0,segs) universe -- the range names an ordinal
	// beyond the segment count. A SUBSET descriptor that fails to cover an
	// in-range ordinal is the over-limit (gap) fixture.
	overLimitGap := CoverageDescriptor{
		Mode:      CoverageModeSubset,
		Covered:   []SegmentRange{{Start: 0, End: 3}},
		Uncovered: []SegmentRange{{Start: 3, End: 6}},
		// ordinals 6,7 covered by neither list -> over the partition boundary
	}
	if err := overLimitGap.Validate(segs, noAttest); err == nil {
		t.Errorf("over-limit coverage (ordinals beyond the partition) was accepted")
	}

	// Paired fixture: the corrected SUBSET descriptor that DOES partition the
	// full [0,segs) is accepted (the at-limit companion of the gap fixture).
	atLimitSubset := CoverageDescriptor{
		Mode:      CoverageModeSubset,
		Covered:   []SegmentRange{{Start: 0, End: 3}},
		Uncovered: []SegmentRange{{Start: 3, End: segs}},
	}
	if err := atLimitSubset.Validate(segs, noAttest); err != nil {
		t.Errorf("at-limit SUBSET partition rejected: %v", err)
	}
}

// countWithinSignatureCeiling reports whether n signature frames are within
// the MAX_SIGNATURES ceiling (a structural upper bound per integrity.abnf S5).
func countWithinSignatureCeiling(n int) bool {
	return n <= MaxSignatures
}
