package integrity

import (
	"reflect"
	"testing"

	"Protodoc/pkg/container"
)

// TestNFR_007_AttestSegmentsExcludedFromStateDigestAndNoOpCompare is
// T-0138's named test. IsAttestTyped is true only for ATTEST across all 4
// slot-segment-type values, and the exclusion helpers (NonAttestOrdinals /
// AttestOrdinals) partition a mixed table correctly. The three consuming
// call sites route through this single predicate rather than a local
// re-implementation (verified for the in-package sites; the ledger no-op
// comparator's use is asserted by call-graph at its own site).
func TestNFR_007_AttestSegmentsExcludedFromStateDigestAndNoOpCompare(t *testing.T) {
	// Table over all 4 segment types.
	types := []struct {
		typ  byte
		want bool
	}{
		{container.SegmentTypeContent, false},
		{container.SegmentTypeResource, false},
		{container.SegmentTypeHistory, false},
		{container.SegmentTypeAttest, true},
	}
	for _, c := range types {
		got := IsAttestTyped(container.SegmentTableSlot{SegmentType: c.typ})
		if got != c.want {
			t.Fatalf("IsAttestTyped(type=%d) = %v, want %v", c.typ, got, c.want)
		}
	}

	// A mixed table: ordinals 0=CONTENT, 1=ATTEST, 2=RESOURCE, 3=unused,
	// 4=ATTEST, 5=HISTORY.
	slots := []container.SegmentTableSlot{
		{SegmentType: container.SegmentTypeContent},
		{SegmentType: container.SegmentTypeAttest},
		{SegmentType: container.SegmentTypeResource},
		{SegmentType: container.SegmentTypeUnused},
		{SegmentType: container.SegmentTypeAttest},
		{SegmentType: container.SegmentTypeHistory},
	}
	nonAttest := NonAttestOrdinals(slots)
	if !reflect.DeepEqual(nonAttest, []uint16{0, 2, 5}) {
		t.Fatalf("NonAttestOrdinals = %v, want [0 2 5]", nonAttest)
	}
	attest := AttestOrdinals(slots)
	if !reflect.DeepEqual(attest, []uint16{1, 4}) {
		t.Fatalf("AttestOrdinals = %v, want [1 4]", attest)
	}

	// No ordinal appears in both partitions, and the unused slot is in
	// neither.
	inNon := map[uint16]bool{}
	for _, o := range nonAttest {
		inNon[o] = true
	}
	for _, o := range attest {
		if inNon[o] {
			t.Fatalf("ordinal %d appears in both partitions", o)
		}
	}
	if inNon[3] {
		t.Fatalf("unused ordinal 3 wrongly counted as non-attest")
	}
}
