package merge

import (
	"errors"
	"math"
	"testing"

	"Protodoc/pkg/history"
)

// TestCON_024_CON_025_GuardsAtDeclaredWireLimit is T-0364's named conformance
// test (CON-024, CON-025). It pins the retention-point and history-mode guards
// exactly at their documented boundaries: an at-limit fixture must pass, and an
// over-limit-by-one fixture must be refused with the SAME named-boundary error
// shape T-0236/T-0237 already produce. Boundary cases only; the clearly-valid
// and clearly-invalid cases live in the T-0236/T-0237 tests. (This task does
// not claim CON-010; it uses the guard requirements this milestone owns.)
func TestCON_024_CON_025_GuardsAtDeclaredWireLimit(t *testing.T) {
	// --- CON-024 retention-point boundary ---
	// The divergence point is exactly the retention point: an incoming change
	// AT the retention point is retained (permitted); a change ONE ordinal
	// BEFORE it is pre-trim (refused). Test at the wire maximum (uint16 max).
	for _, rp := range []uint16{0 + 1, 100, math.MaxUint16} {
		decl, err := history.NewDeclarationWithRetentionPoint(history.ModeRetainedFromPoint, rp)
		if err != nil {
			t.Fatalf("declare retention point %d: %v", rp, err)
		}

		// At limit: a change exactly AT the retention point -> permitted.
		if err := CheckRetentionPoint(decl, []uint16{rp}); err != nil {
			t.Errorf("rp=%d: change at the retention point must pass, got %v", rp, err)
		}

		// Over limit by one: a change ONE ordinal before -> refused with the
		// same named-boundary error shape.
		overErr := CheckRetentionPoint(decl, []uint16{rp - 1})
		if !errors.Is(overErr, ErrMergeAcrossRetentionPoint) {
			t.Errorf("rp=%d: change one before must be refused, got %v", rp, overErr)
		}
		var re *RetentionPointError
		if !errors.As(overErr, &re) || re.RetentionPoint != rp || re.OffendingOrdinal != rp-1 {
			t.Errorf("rp=%d: refusal must name the boundary (rp=%d, offending=%d), got %v", rp, rp, rp-1, overErr)
		}
	}

	// --- CON-025 history-mode boundary ---
	// The "limit" is a single history-mode field: identical modes on both
	// branches (the maximal consistent set) pass; a single differing field is
	// refused naming both modes. Exercise the enum's extreme value too.
	modes := []history.Mode{history.ModeComplete, history.ModeRetainedFromPoint, history.ModeNone}
	for _, m := range modes {
		var a history.Declaration
		var err error
		if m == history.ModeRetainedFromPoint {
			a, err = history.NewDeclarationWithRetentionPoint(m, math.MaxUint16)
		} else {
			a, err = history.NewDeclaration(m)
		}
		if err != nil {
			t.Fatalf("declare mode %s: %v", history.ModeName(m), err)
		}
		// At limit: identical single mode value on both branches -> passes.
		if err := CheckHistoryModeMatch(a, a); err != nil {
			t.Errorf("mode %s: identical modes must pass, got %v", history.ModeName(m), err)
		}
	}

	// Over limit by one: a single differing mode field between the branches ->
	// refused naming both modes.
	c, _ := history.NewDeclaration(history.ModeComplete)
	n, _ := history.NewDeclaration(history.ModeNone)
	mismatch := CheckHistoryModeMatch(c, n)
	if !errors.Is(mismatch, ErrHistoryModeMismatch) {
		t.Errorf("single differing mode must be refused, got %v", mismatch)
	}
	var me *HistoryModeMismatchError
	if !errors.As(mismatch, &me) || me.ModeA != history.ModeComplete || me.ModeB != history.ModeNone {
		t.Errorf("refusal must name both boundary modes, got %v", mismatch)
	}
}
