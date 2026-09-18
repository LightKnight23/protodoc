package migrate

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestCON_016_RescindResignRecordConformanceVectors is T-0305's named
// conformance test (CON-016 / FR-110). It ships a valid at-limit fixture, a
// duplicate-identifier pair (rejected per FR-110 naming both locations), and a
// one-past-limit fixture (rejected).
func TestCON_016_RescindResignRecordConformanceVectors(t *testing.T) {
	var prior pdlfmt.UnitID
	prior[0] = 0x9
	var so, state pdlfmt.Digest256
	so[0] = 0x5
	state[0] = 0x6
	var sigv [64]byte
	sigv[0] = 0x7

	rec := RescindResignRecord{
		PriorSignatureRef: prior,
		NewParamSetID:     1,
		NewSignedObject:   so,
		NewSignatureValue: sigv,
		MigrationStateID:  state,
	}

	// Valid at-limit fixture: encodes to exactly the max size and round-trips.
	enc := EncodeRescindResign(nil, rec)
	if len(enc) != RescindResignWireMax {
		t.Fatalf("encoded length %d, want %d", len(enc), RescindResignWireMax)
	}
	got, n, err := DecodeRescindResign(enc)
	if err != nil || n != RescindResignWireMax || got != rec {
		t.Fatalf("roundtrip: got %+v (n=%d) err=%v", got, n, err)
	}

	// One-past-limit fixture: one extra octet is rejected.
	over := append(append([]byte(nil), enc...), 0x00)
	if _, _, err := DecodeRescindResign(over); err == nil {
		t.Errorf("one-past-limit buffer (%d octets) must be rejected", len(over))
	}
	// One-below-limit is also rejected (exact length required).
	if _, _, err := DecodeRescindResign(enc[:len(enc)-1]); err == nil {
		t.Errorf("one-below-limit buffer must be rejected")
	}

	// Duplicate-identifier pair: two records with the same (migration_state_id,
	// new_param_set) at distinct locations -> rejected naming BOTH locations.
	dupA := PlacedRescindResign{Record: rec, Location: RRLocation{SegmentOrdinal: 0, IntraOffset: 0}}
	dupB := PlacedRescindResign{Record: rec, Location: RRLocation{SegmentOrdinal: 2, IntraOffset: 146}}
	err = CheckDuplicateRescindResign([]PlacedRescindResign{dupA, dupB})
	var de *DuplicateRRError
	if !errors.As(err, &de) {
		t.Fatalf("duplicate identity: err = %v, want DuplicateRRError", err)
	}
	if de.First != dupA.Location || de.Later != dupB.Location {
		t.Errorf("duplicate must name both locations, got first=%s later=%s", de.First, de.Later)
	}

	// Distinct identities pass.
	other := rec
	other.NewParamSetID = 1
	other.MigrationStateID[0] = 0xAB
	if err := CheckDuplicateRescindResign([]PlacedRescindResign{
		dupA, {Record: other, Location: RRLocation{SegmentOrdinal: 3, IntraOffset: 0}},
	}); err != nil {
		t.Errorf("distinct identities should pass, got %v", err)
	}
}
