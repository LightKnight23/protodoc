package validate

import (
	"errors"
	"strings"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_110_DuplicateIdentifierNamesBothLocations is T-0120's named
// conformance test. Two RUN records sharing a run_id, and two
// SegmentTableSlot entries sharing a unit id, both reject naming both
// locations; a single-instance document is unaffected.
func TestFR_110_DuplicateIdentifierNamesBothLocations(t *testing.T) {
	id := func(b byte) pdlfmt.UnitID { var u pdlfmt.UnitID; u[0] = b; return u }

	// Two RUN records sharing a run_id at distinct locations.
	dupRuns := []IdentifiedUnit{
		{ID: id(1), Location: Location{SegmentOrdinal: 0, IntraOffset: 64}},
		{ID: id(2), Location: Location{SegmentOrdinal: 0, IntraOffset: 128}},
		{ID: id(1), Location: Location{SegmentOrdinal: 3, IntraOffset: 200}}, // duplicate of id(1)
	}
	err := CheckDuplicateIdentifiers(dupRuns)
	var de *DuplicateIdentifierError
	if !errors.As(err, &de) {
		t.Fatalf("duplicate run_id: got %v, want DuplicateIdentifierError", err)
	}
	if de.ID != id(1) {
		t.Fatalf("duplicate error names id %x, want id(1)", de.ID)
	}
	// Both locations named, and both appear in the message.
	if de.First != (Location{0, 64}) || de.Later != (Location{3, 200}) {
		t.Fatalf("duplicate error locations = %v / %v, want {0,64} / {3,200}", de.First, de.Later)
	}
	msg := de.Error()
	if !strings.Contains(msg, "segment 0 offset 64") || !strings.Contains(msg, "segment 3 offset 200") {
		t.Fatalf("message does not name both locations: %q", msg)
	}

	// Two index entries (SegmentTableSlot-level) sharing a unit id.
	dupIndex := []IdentifiedUnit{
		{ID: id(5), Location: Location{SegmentOrdinal: 5, IntraOffset: 0}},
		{ID: id(5), Location: Location{SegmentOrdinal: 9, IntraOffset: 0}},
	}
	if err := CheckDuplicateIdentifiers(dupIndex); !errors.As(err, &de) {
		t.Fatalf("duplicate index entry: got %v, want DuplicateIdentifierError", err)
	}

	// A single-instance document (all ids unique) passes.
	unique := []IdentifiedUnit{
		{ID: id(1), Location: Location{SegmentOrdinal: 0, IntraOffset: 0}},
		{ID: id(2), Location: Location{SegmentOrdinal: 0, IntraOffset: 64}},
		{ID: id(3), Location: Location{SegmentOrdinal: 1, IntraOffset: 0}},
	}
	if err := CheckDuplicateIdentifiers(unique); err != nil {
		t.Fatalf("unique-identifier document rejected: %v", err)
	}
}
