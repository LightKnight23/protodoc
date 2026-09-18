package merge

import (
	"testing"

	"Protodoc/pkg/history"
)

// TestFR_092_DeleteDominatesConcurrentOps is T-0231's named unit test
// (FR-092). On the same target, a DELETE dominates any concurrent non-delete
// (R3 DELETE-DOMINATES): a delete concurrent with a value/position claim
// removes the unit, in any delivery order. Absent a delete, the R2 tiebreak
// selects the largest-state-id value.
func TestFR_092_DeleteDominatesConcurrentOps(t *testing.T) {
	sid := func(b byte) [32]byte { var s [32]byte; s[0] = b; return s }
	target := mTarget(0x01)
	base := Doc{target: []byte("original")}

	// A concurrent delete and value-claim on the same target: delete wins.
	del := Op{Kind: history.OpDelete, Target: target, StateID: sid(0x10)}
	set := Op{Kind: history.OpValueClaim, Target: target, StateID: sid(0x99), Value: []byte("edited")}

	// The classifier says a same-target pair involving a delete is
	// DELETE-DOMINATES regardless of the other kind.
	for _, other := range []history.OpKind{history.OpValueClaim, history.OpSequencePositionClaim, history.OpDelete} {
		o, _ := Classify(history.OpDelete, other, Same)
		if o != DeleteDominates {
			t.Errorf("delete vs %v on same target = %v, want DeleteDominates", other, o)
		}
		o2, _ := Classify(other, history.OpDelete, Same)
		if o2 != DeleteDominates {
			t.Errorf("%v vs delete on same target = %v, want DeleteDominates", other, o2)
		}
	}

	// Apply resolves to the target removed, in either delivery order, even
	// though the value-claim has the LARGER state-id (delete still dominates).
	for _, order := range [][]Op{{del, set}, {set, del}} {
		got := ResolveSameTarget(base, target, order)
		if _, present := got[target]; present {
			t.Errorf("delete did not dominate: target still present after %v", order)
		}
	}

	// Without a delete, the largest-state-id value-claim wins (R2).
	set1 := Op{Kind: history.OpValueClaim, Target: target, StateID: sid(0x10), Value: []byte("v1")}
	set2 := Op{Kind: history.OpValueClaim, Target: target, StateID: sid(0x20), Value: []byte("v2")}
	got := ResolveSameTarget(base, target, []Op{set1, set2})
	if string(got[target]) != "v2" {
		t.Errorf("R2: value = %q, want v2 (larger state-id wins)", got[target])
	}
	// Order-independent.
	got2 := ResolveSameTarget(base, target, []Op{set2, set1})
	if string(got2[target]) != string(got[target]) {
		t.Error("same-target resolution is order-dependent")
	}
}
