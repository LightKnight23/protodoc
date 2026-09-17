package integrity

import (
	"bytes"
	"errors"
	"testing"
)

// TestFR_002_CoverageDescriptorCanonicalImplementationRejectsEachPDCoverViolation
// is T-0139's named test. A TOTAL-mode descriptor covering every non-ATTEST
// ordinal validates; a SUBSET-mode descriptor with an unmerged-adjacent
// pair, a zero-length range, a gap, an overlap, or an ATTEST ordinal named
// each fail with a distinct named reason; and encode/decode round-trips.
func TestFR_002_CoverageDescriptorCanonicalImplementationRejectsEachPDCoverViolation(t *testing.T) {
	// Document: 5 ordinals, ordinal 3 is ATTEST.
	const segmentCount = 5
	attest := map[uint16]bool{3: true}

	// TOTAL mode: covers all non-ATTEST (0,1,2,4), uncovered empty. Valid.
	total := CoverageDescriptor{
		Mode:      CoverageModeTotal,
		Covered:   []SegmentRange{{Start: 0, End: 3}, {Start: 4, End: 5}}, // omits attest ordinal 3
		Uncovered: nil,
		Bitmask:   CoverageBitHeader,
	}
	if err := total.Validate(segmentCount, attest); err != nil {
		t.Fatalf("valid TOTAL descriptor rejected: %v", err)
	}

	// A well-formed SUBSET: covered {0,1}, uncovered {2,4}, ordinal 3 attest.
	validSubset := CoverageDescriptor{
		Mode:      CoverageModeSubset,
		Covered:   []SegmentRange{{Start: 0, End: 2}},
		Uncovered: []SegmentRange{{Start: 2, End: 3}, {Start: 4, End: 5}},
		Bitmask:   0,
	}
	if err := validSubset.Validate(segmentCount, attest); err != nil {
		t.Fatalf("valid SUBSET descriptor rejected: %v", err)
	}

	// Round-trip encode/decode of the valid subset.
	enc, err := validSubset.Encode(nil)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, n, err := DecodeCoverageDescriptor(enc)
	if err != nil || n != len(enc) {
		t.Fatalf("DecodeCoverageDescriptor: err=%v consumed=%d/%d", err, n, len(enc))
	}
	reEnc, _ := got.Encode(nil)
	if !bytes.Equal(enc, reEnc) {
		t.Fatalf("coverage descriptor did not round-trip byte-exact")
	}

	// PD-COVER-001: zero-length range.
	zeroLen := CoverageDescriptor{Mode: CoverageModeSubset, Covered: []SegmentRange{{Start: 1, End: 1}}}
	if err := zeroLen.Validate(segmentCount, attest); !errors.Is(err, ErrCoverZeroLengthRange) {
		t.Fatalf("zero-length range: got %v, want ErrCoverZeroLengthRange", err)
	}

	// PD-COVER-002: unmerged adjacent pair ([0,1)+[1,2) should be [0,2)).
	adjacent := CoverageDescriptor{Mode: CoverageModeSubset, Covered: []SegmentRange{{Start: 0, End: 1}, {Start: 1, End: 2}}, Uncovered: []SegmentRange{{Start: 2, End: 5}}}
	if err := adjacent.Validate(segmentCount, attest); !errors.Is(err, ErrCoverNotCanonical) {
		t.Fatalf("unmerged adjacent: got %v, want ErrCoverNotCanonical", err)
	}
	// PD-COVER-002: overlap.
	overlap := CoverageDescriptor{Mode: CoverageModeSubset, Covered: []SegmentRange{{Start: 0, End: 3}, {Start: 2, End: 4}}}
	if err := overlap.Validate(segmentCount, attest); !errors.Is(err, ErrCoverNotCanonical) {
		t.Fatalf("overlap: got %v, want ErrCoverNotCanonical", err)
	}

	// PD-COVER-003: gap -- ordinal 1 in neither list (SUBSET).
	gap := CoverageDescriptor{Mode: CoverageModeSubset, Covered: []SegmentRange{{Start: 0, End: 1}}, Uncovered: []SegmentRange{{Start: 2, End: 3}, {Start: 4, End: 5}}}
	if err := gap.Validate(segmentCount, attest); !errors.Is(err, ErrCoverPartition) {
		t.Fatalf("gap: got %v, want ErrCoverPartition", err)
	}

	// PD-COVER-004: an ATTEST ordinal (3) named in a list.
	namesAttest := CoverageDescriptor{Mode: CoverageModeSubset, Covered: []SegmentRange{{Start: 0, End: 4}}, Uncovered: []SegmentRange{{Start: 4, End: 5}}}
	if err := namesAttest.Validate(segmentCount, attest); !errors.Is(err, ErrCoverAttestNamed) {
		t.Fatalf("names attest ordinal: got %v, want ErrCoverAttestNamed", err)
	}

	// Invalid mode and reserved bits.
	if err := (CoverageDescriptor{Mode: 2}).Validate(segmentCount, attest); !errors.Is(err, ErrCoverInvalidMode) {
		t.Fatalf("invalid mode: got %v, want ErrCoverInvalidMode", err)
	}
	if err := (CoverageDescriptor{Mode: CoverageModeSubset, Bitmask: 0x20}).Validate(segmentCount, attest); !errors.Is(err, ErrCoverReservedBits) {
		t.Fatalf("reserved bits: got %v, want ErrCoverReservedBits", err)
	}
}
