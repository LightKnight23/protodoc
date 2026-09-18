package merge

import "testing"

// TestFR_094_MoveCycleRetainsSmallerStateId is T-0230's named unit test
// (FR-094). Concurrent move operations that would place a unit within its own
// subtree are resolved by retaining the move whose originating state-id is
// lexicographically SMALLER, recording the other as a conflict against the
// moved unit's identity, and leaving that unit's subtree at its pre-move
// parent.
func TestFR_094_MoveCycleRetainsSmallerStateId(t *testing.T) {
	sid := func(b byte) [32]byte { var s [32]byte; s[0] = b; return s }
	unitX := mTarget(0x01)
	unitY := mTarget(0x02)
	preParentX := mTarget(0xF0)
	preParentY := mTarget(0xF1)

	// m1 (state 0x10) moves X under Y; m2 (state 0x20) moves Y under X --
	// together a cycle. m1 has the smaller state-id, so it is retained.
	m1 := Move{Unit: unitX, NewParent: unitY, StateID: sid(0x10)}
	m2 := Move{Unit: unitY, NewParent: unitX, StateID: sid(0x20)}

	retained, restore, conflict := ResolveMoveCycle(m1, m2, preParentY)
	if retained.StateID != m1.StateID {
		t.Errorf("retained move state-id = %x, want the smaller %x", retained.StateID, m1.StateID)
	}
	// The rejected move (m2) is recorded as a conflict against its moved
	// unit's identity (Y).
	if conflict.Unit != unitY {
		t.Errorf("conflict recorded against %x, want the moved unit Y %x", conflict.Unit, unitY)
	}
	if conflict.RejectedMove.StateID != m2.StateID {
		t.Errorf("rejected move state-id = %x, want %x", conflict.RejectedMove.StateID, m2.StateID)
	}
	if conflict.RetainedMove.StateID != m1.StateID {
		t.Errorf("conflict's retained move should be m1")
	}
	// Y's subtree is left at its pre-move parent.
	if restore != preParentY {
		t.Errorf("losing unit restore parent = %x, want pre-move parent %x", restore, preParentY)
	}

	// Symmetric: if m2 has the smaller id, m2 is retained and m1 conflicts.
	m1b := Move{Unit: unitX, NewParent: unitY, StateID: sid(0x30)}
	m2b := Move{Unit: unitY, NewParent: unitX, StateID: sid(0x05)}
	ret2, _, conf2 := ResolveMoveCycle(m1b, m2b, preParentX)
	if ret2.StateID != m2b.StateID {
		t.Errorf("retained = %x, want the smaller %x", ret2.StateID, m2b.StateID)
	}
	if conf2.Unit != unitX {
		t.Errorf("conflict recorded against %x, want X %x", conf2.Unit, unitX)
	}

	// Deterministic: same inputs, same resolution.
	r3, _, _ := ResolveMoveCycle(m1, m2, preParentY)
	if r3.StateID != retained.StateID {
		t.Error("ResolveMoveCycle is not deterministic")
	}
}
