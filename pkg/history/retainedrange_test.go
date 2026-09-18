package history

import (
	"testing"

	"Protodoc/pkg/container"
)

// TestFR_060_RetainedFromPointReconstructsGuaranteedRange is T-0214's named
// integration test (FR-060). In history-retained-from-point mode, every state
// published AT OR AFTER the declared retention point is reconstructable
// octet-for-octet from the current file; states before it carry no guarantee.
func TestFR_060_RetainedFromPointReconstructsGuaranteedRange(t *testing.T) {
	const retentionPoint uint16 = 3
	d, err := NewDeclarationWithRetentionPoint(ModeRetainedFromPoint, retentionPoint)
	if err != nil {
		t.Fatalf("declaration: %v", err)
	}

	// States by ordinal: 1,2 are before the retention point; 3,4,5 at/after.
	for ord := uint16(0); ord < 6; ord++ {
		guaranteed := StateReconstructable(d.RetentionPoint(), ord)
		wantGuaranteed := ord >= retentionPoint
		if guaranteed != wantGuaranteed {
			t.Errorf("ordinal %d: reconstructable=%v, want %v", ord, guaranteed, wantGuaranteed)
		}
	}

	// The retained range (ordinals >= retentionPoint) reconstructs
	// octet-for-octet from the operation log for those states.
	u := hUnitID(0x01)
	// Ops authored at states s3,s4,s5 (the retained range).
	s3, s4, s5 := hStateID(0x03), hStateID(0x04), hStateID(0x05)
	ops := []OperationRecord{
		{Kind: OpValueClaim, Target: u, OrderKey: s3, Predecessor: container.StateID{}, Payload: []byte("v3")},
		{Kind: OpValueClaim, Target: u, OrderKey: s4, Predecessor: s3, Payload: []byte("v4")},
		{Kind: OpValueClaim, Target: u, OrderKey: s5, Predecessor: s4, Payload: []byte("v5")},
	}
	base := newSnapshot()
	for _, target := range []container.StateID{s3, s4, s5} {
		snap, err := Reconstruct(base, ops, target)
		if err != nil {
			t.Fatalf("Reconstruct(%x) in retained range: %v", target, err)
		}
		snap2, _ := Reconstruct(base, ops, target)
		if string(snap.Octets()) != string(snap2.Octets()) {
			t.Errorf("state %x not deterministically reconstructed", target)
		}
	}

	// Exactly-at-the-point ordinal is guaranteed (closed lower bound).
	if !StateReconstructable(retentionPoint, retentionPoint) {
		t.Error("the state exactly at the retention point must be reconstructable")
	}
	// One before is not guaranteed.
	if StateReconstructable(retentionPoint, retentionPoint-1) {
		t.Error("a state before the retention point must carry no reconstruction guarantee")
	}
}
