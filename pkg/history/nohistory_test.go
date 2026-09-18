package history

import "testing"

// TestCON_022_NoHistoryModeRetainsOnlyCurrentState is T-0215's named
// integration test (CON-022). A no-history document retains ONLY the current
// state: no prior published state is reconstructable, and (unlike
// retained-from-point) there is no retention point that would make any earlier
// ordinal reconstructable.
func TestCON_022_NoHistoryModeRetainsOnlyCurrentState(t *testing.T) {
	d, err := NewDeclaration(ModeNone)
	if err != nil {
		t.Fatalf("NewDeclaration(ModeNone): %v", err)
	}
	if d.Mode() != ModeNone {
		t.Fatalf("mode = %v, want no-history", d.Mode())
	}

	// No-history has no meaningful retention point.
	if d.HasMeaningfulRetentionPoint() {
		t.Error("no-history mode must have no meaningful retention point")
	}

	// Model "only the current state retained" as a retention point at the
	// current ordinal: nothing before the current state is reconstructable.
	const current uint16 = 100
	// Under no-history the retained range is exactly {current}: every earlier
	// ordinal is not reconstructable.
	for ord := uint16(0); ord < current; ord++ {
		if StateReconstructable(current, ord) {
			t.Errorf("no-history: ordinal %d before current must not be reconstructable", ord)
		}
	}
	// The current state itself is available.
	if !StateReconstructable(current, current) {
		t.Error("no-history: the current state must be available")
	}

	// A no-history document carries no operation-log history to replay: an
	// empty op log yields only the current (base) snapshot, and no prior
	// target reconstructs.
	base := newSnapshot()
	base.units[hUnitID(0x01)] = []byte("current content")
	if _, err := Reconstruct(base, nil, hStateID(0x05)); err == nil {
		t.Error("no-history: reconstructing a prior state from an empty log should error")
	}
	// The current snapshot's octets are stable.
	if string(base.Octets()) != string(base.clone().Octets()) {
		t.Error("current snapshot octets not stable")
	}
}
