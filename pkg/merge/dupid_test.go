package merge

import (
	"errors"
	"testing"
)

// TestFR_024_RefuseMergeOnDuplicateIdentifier is T-0233's named unit test
// (FR-024). IF a merge input pair contains two DISTINCT content units carrying
// the same identifier, the merge is refused, naming both units and their
// source documents. A shared id with identical content (the same unit / a
// common ancestor) is not a collision and is permitted.
func TestFR_024_RefuseMergeOnDuplicateIdentifier(t *testing.T) {
	id := mTarget(0x01)
	dgA := [32]byte{0xAA}
	dgB := [32]byte{0xBB}

	// Distinct units (different digests) sharing an id -> refuse.
	a := []MergeUnit{{ID: id, Digest: dgA, Source: "doc-A.pdl"}}
	b := []MergeUnit{{ID: id, Digest: dgB, Source: "doc-B.pdl"}}
	err := CheckDuplicateIdentifier(a, b)
	if !errors.Is(err, ErrDuplicateIdentifier) {
		t.Fatalf("distinct units sharing id: err = %v, want ErrDuplicateIdentifier", err)
	}
	var de *DuplicateIdentifierError
	if !errors.As(err, &de) {
		t.Fatalf("expected *DuplicateIdentifierError, got %T", err)
	}
	if de.ID != id {
		t.Errorf("error names id %x, want %x", de.ID, id)
	}
	// Both source documents are named.
	if de.SourceA != "doc-A.pdl" || de.SourceB != "doc-B.pdl" {
		t.Errorf("error must name both sources, got %q and %q", de.SourceA, de.SourceB)
	}

	// Same id, same content (the same unit / common ancestor) -> permitted.
	same := []MergeUnit{{ID: id, Digest: dgA, Source: "doc-B.pdl"}}
	if err := CheckDuplicateIdentifier(a, same); err != nil {
		t.Errorf("shared id with identical content should be permitted, got %v", err)
	}

	// Disjoint ids -> permitted.
	other := []MergeUnit{{ID: mTarget(0x02), Digest: dgB, Source: "doc-B.pdl"}}
	if err := CheckDuplicateIdentifier(a, other); err != nil {
		t.Errorf("disjoint ids should be permitted, got %v", err)
	}
}
