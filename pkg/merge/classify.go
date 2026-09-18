// Package merge implements Protodoc's concurrent-edit merge: the DP-015
// exhaustive 4-way classification of every pair of concurrent operation kinds,
// Fugue's non-interleaving sequence order, and the conflict-resolution rules.
// Every operation is exhaustively a position-claim, a value-claim, or a delete
// (history.OpKind), so classifying a concurrent pair by its two kinds covers
// every case with no wildcard and no implementation-defined cell (FR-092,
// HC-026). It builds on pkg/history (op kinds, state ids) and pkg/container.
package merge

import (
	"fmt"

	"Protodoc/pkg/history"
)

// Outcome is one of the DP-015 classification's exactly-four outcomes for a
// concurrent operation pair.
type Outcome int

const (
	// DisjointCommute (0): the two operations touch disjoint targets and
	// commute; either delivery order yields the identical result.
	DisjointCommute Outcome = iota
	// SequenceOrder (1): R1 -- both claim a position in the same sequence;
	// resolved by Fugue's proven non-interleaving order.
	SequenceOrder
	// TotalOrderTiebreak (2): R2 -- both set a value on the same target;
	// resolved by the unsigned big-endian state_id comparison (last-writer by
	// state id, deterministically).
	TotalOrderTiebreak
	// DeleteDominates (3): R3 -- at least one operation is a delete; the
	// delete dominates a concurrent non-delete on the same target.
	DeleteDominates
)

func (o Outcome) String() string {
	switch o {
	case DisjointCommute:
		return "DISJOINT-COMMUTE"
	case SequenceOrder:
		return "R1-SEQUENCE-ORDER"
	case TotalOrderTiebreak:
		return "R2-TOTAL-ORDER-TIEBREAK"
	case DeleteDominates:
		return "R3-DELETE-DOMINATES"
	default:
		return "invalid"
	}
}

// SameTarget indicates whether the two concurrent operations act on the same
// target unit. Disjoint targets commute regardless of kind; same-target pairs
// are classified by their kinds.
type SameTarget bool

const (
	Disjoint SameTarget = false
	Same     SameTarget = true
)

// ErrInvalidOpKind is returned by Classify for an op-kind outside the closed
// three-value set (a defensive check; the enum itself is closed).
var ErrInvalidOpKind = fmt.Errorf("merge: operation kind outside the closed set {position-claim, value-claim, delete}")

// Classify returns the single DP-015 outcome for a concurrent pair of
// operations of kinds a and b acting on the same-or-disjoint target. The
// dispatch is EXHAUSTIVE over the closed 3-kind set with NO wildcard/default
// branch: every (a, b) pair maps to exactly one outcome, so two conforming
// implementations classify every concurrent pair identically with zero
// implementation-defined cells (FR-092). Disjoint-target pairs always
// commute; same-target pairs dispatch by kind.
func Classify(a, b history.OpKind, target SameTarget) (Outcome, error) {
	if !kindValid(a) || !kindValid(b) {
		return 0, ErrInvalidOpKind
	}
	if target == Disjoint {
		return DisjointCommute, nil
	}
	// Same-target dispatch over the closed 3x3 kind matrix. A delete on either
	// side dominates (R3). Two position-claims order by Fugue (R1). Any pair
	// involving a value-claim (and no delete) is a value contention resolved
	// by the R2 state-id tiebreak.
	switch {
	case a == history.OpDelete || b == history.OpDelete:
		return DeleteDominates, nil
	case a == history.OpSequencePositionClaim && b == history.OpSequencePositionClaim:
		return SequenceOrder, nil
	default:
		// The remaining same-target, delete-free pairs all involve at least
		// one value-claim: (value,value), (value,position), (position,value).
		// This branch is reached ONLY for those enumerated cases -- it is not a
		// wildcard over unknown kinds (kindValid rejected those above) but the
		// exhaustive completion of the closed 3x3 matrix.
		return TotalOrderTiebreak, nil
	}
}

func kindValid(k history.OpKind) bool {
	return k == history.OpSequencePositionClaim || k == history.OpValueClaim || k == history.OpDelete
}
