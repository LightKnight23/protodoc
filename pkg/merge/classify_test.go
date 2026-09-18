package merge

import (
	"errors"
	"testing"

	"Protodoc/pkg/history"
)

// TestFR_092_ClassifyDispatchHasNoWildcardCase is T-0225's named unit test
// (FR-092). The classifier defines exactly one outcome for every pair of
// concurrent operation kinds over the CLOSED 3-kind set, with zero cases left
// implementation-defined and NO wildcard over unknown kinds: an op-kind outside
// the closed set is REJECTED (not silently classified), and every in-set pair
// yields exactly one defined outcome.
func TestFR_092_ClassifyDispatchHasNoWildcardCase(t *testing.T) {
	kinds := []history.OpKind{history.OpSequencePositionClaim, history.OpValueClaim, history.OpDelete}

	// Every in-set pair (both same-target and disjoint) yields exactly one
	// defined outcome, no error.
	for _, a := range kinds {
		for _, b := range kinds {
			for _, tgt := range []SameTarget{Same, Disjoint} {
				o, err := Classify(a, b, tgt)
				if err != nil {
					t.Errorf("Classify(%v,%v,%v) errored on in-set kinds: %v", a, b, tgt, err)
				}
				if o < DisjointCommute || o > DeleteDominates {
					t.Errorf("Classify(%v,%v,%v) = out-of-range outcome %d", a, b, tgt, o)
				}
				if o.String() == "invalid" {
					t.Errorf("Classify(%v,%v,%v) produced an invalid outcome", a, b, tgt)
				}
			}
		}
	}

	// An out-of-set op-kind is rejected, never classified by a wildcard.
	for _, bad := range []history.OpKind{0x03, 0x7F, 0xFF} {
		if _, err := Classify(bad, history.OpValueClaim, Same); !errors.Is(err, ErrInvalidOpKind) {
			t.Errorf("Classify(bad=0x%02x): err = %v, want ErrInvalidOpKind", uint8(bad), err)
		}
		if _, err := Classify(history.OpValueClaim, bad, Same); !errors.Is(err, ErrInvalidOpKind) {
			t.Errorf("Classify(_, bad=0x%02x): err = %v, want ErrInvalidOpKind", uint8(bad), err)
		}
	}

	// The classifier's source has no `default`-that-produces-an-outcome over
	// unknown kinds: the only default branch is the exhaustive completion of
	// the closed matrix, reached solely for delete-free value pairs. We verify
	// that completion is exactly TotalOrderTiebreak for those pairs.
	for _, pair := range [][2]history.OpKind{
		{history.OpValueClaim, history.OpValueClaim},
		{history.OpValueClaim, history.OpSequencePositionClaim},
		{history.OpSequencePositionClaim, history.OpValueClaim},
	} {
		o, _ := Classify(pair[0], pair[1], Same)
		if o != TotalOrderTiebreak {
			t.Errorf("delete-free value pair (%v,%v) = %v, want R2 TotalOrderTiebreak", pair[0], pair[1], o)
		}
	}
}
