package merge

import (
	"testing"

	"Protodoc/pkg/history"
)

// TestFR_092_ExhaustiveOperationKindPairClassification is T-0232's named
// conformance test (FR-092). It enumerates the FULL classification matrix --
// every ordered pair of the closed 3 op-kinds under both same-target and
// disjoint-target -- and asserts each cell has exactly one defined outcome
// (zero implementation-defined cells), the disjoint cells all commute, the
// same-target cells match the DP-015 rules, and the classification is
// symmetric (Classify(a,b) == Classify(b,a)) so delivery order does not change
// the outcome.
func TestFR_092_ExhaustiveOperationKindPairClassification(t *testing.T) {
	kinds := []history.OpKind{history.OpSequencePositionClaim, history.OpValueClaim, history.OpDelete}

	// Expected same-target outcome per (a,b) kind pair.
	expectSame := func(a, b history.OpKind) Outcome {
		switch {
		case a == history.OpDelete || b == history.OpDelete:
			return DeleteDominates
		case a == history.OpSequencePositionClaim && b == history.OpSequencePositionClaim:
			return SequenceOrder
		default:
			return TotalOrderTiebreak
		}
	}

	covered := 0
	for _, a := range kinds {
		for _, b := range kinds {
			// Disjoint: always commute.
			if o, err := Classify(a, b, Disjoint); err != nil || o != DisjointCommute {
				t.Errorf("disjoint (%v,%v) = %v err %v, want DISJOINT-COMMUTE", a, b, o, err)
			}
			covered++

			// Same target: matches the DP-015 rule, exactly one outcome.
			o, err := Classify(a, b, Same)
			if err != nil {
				t.Errorf("same-target (%v,%v) errored: %v", a, b, err)
			}
			if o != expectSame(a, b) {
				t.Errorf("same-target (%v,%v) = %v, want %v", a, b, o, expectSame(a, b))
			}
			covered++

			// Symmetric: order of the two concurrent kinds does not change the
			// outcome.
			rev, _ := Classify(b, a, Same)
			if rev != o {
				t.Errorf("classification not symmetric for (%v,%v): %v vs %v", a, b, o, rev)
			}
		}
	}

	// The full matrix (9 kind pairs x 2 target modes = 18 cells) is covered
	// with no gap.
	if covered != len(kinds)*len(kinds)*2 {
		t.Errorf("covered %d cells, want %d (full 3x3 x {same,disjoint} matrix)", covered, len(kinds)*len(kinds)*2)
	}
}
