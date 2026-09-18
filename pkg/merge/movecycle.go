// Move-cycle resolution (T-0230, FR-094). IF concurrent move operations would
// place a content unit within its own subtree (a cycle), the format retains the
// move whose originating state identifier is lexicographically SMALLER, records
// the other as a conflict against the moved unit's identity, and leaves that
// unit's subtree at its pre-move parent. This is a fixed, deterministic
// resolution: two implementations pick the identical retained move and record
// the identical conflict.
package merge

import (
	"bytes"

	"Protodoc/pkg/pdlfmt"
)

// Move is a move operation: the unit being moved, its new parent, and the
// originating state-id (the FR-094 tiebreak key).
type Move struct {
	Unit      pdlfmt.UnitID
	NewParent pdlfmt.UnitID
	StateID   [32]byte
}

// MoveConflict records a move that was NOT retained because it would have
// formed a cycle: the losing move, recorded against the moved unit's identity.
type MoveConflict struct {
	Unit         pdlfmt.UnitID // the moved unit's identity the conflict is recorded against
	RejectedMove Move          // the move that lost
	RetainedMove Move          // the move that won (smaller state-id)
}

// ResolveMoveCycle resolves two concurrent moves that together would place a
// unit within its own subtree (a cycle). It retains the move with the
// lexicographically SMALLER originating state-id, returns the retained move,
// the pre-move parent to leave the losing move's unit at, and a MoveConflict
// recording the rejected move against the moved unit's identity (FR-094). If
// the two moves share a state-id (not a real concurrent cycle), the first is
// retained deterministically.
func ResolveMoveCycle(m1, m2 Move, preMoveParent pdlfmt.UnitID) (retained Move, restoreParent pdlfmt.UnitID, conflict MoveConflict) {
	// Smaller originating state-id wins.
	if bytes.Compare(m1.StateID[:], m2.StateID[:]) <= 0 {
		retained = m1
		conflict = MoveConflict{Unit: m2.Unit, RejectedMove: m2, RetainedMove: m1}
	} else {
		retained = m2
		conflict = MoveConflict{Unit: m1.Unit, RejectedMove: m1, RetainedMove: m2}
	}
	// The losing move's unit stays at its pre-move parent (its subtree is not
	// relocated).
	restoreParent = preMoveParent
	return retained, restoreParent, conflict
}
