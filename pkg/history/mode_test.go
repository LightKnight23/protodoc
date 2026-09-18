package history

import (
	"errors"
	"testing"
)

// TestCON_022_HistoryModeClosedEnumAndImmutable is T-0207's named unit test
// (CON-022). Every document declares, at creation and immutably, one history
// mode from the CLOSED set {complete, retained-from-point, no-history}. This
// asserts the closed set is exactly those three, an out-of-range value is
// rejected, and a declaration is immutable: redeclaring the same mode is a
// no-op but declaring a different mode is refused naming the constraint.
func TestCON_022_HistoryModeClosedEnumAndImmutable(t *testing.T) {
	// Closed set of exactly three modes with distinct names.
	if len(AllModes) != 3 {
		t.Fatalf("AllModes has %d entries, want exactly 3 (CON-022)", len(AllModes))
	}
	names := map[string]bool{}
	for _, m := range AllModes {
		if !ValidMode(m) {
			t.Errorf("mode %v not recognised as valid", m)
		}
		names[ModeName(m)] = true
	}
	for _, want := range []string{"complete-history", "history-retained-from-point", "no-history"} {
		if !names[want] {
			t.Errorf("closed set missing mode %q", want)
		}
	}

	// An out-of-range mode is rejected, never defaulted.
	for _, bad := range []Mode{3, 7, 255} {
		if ValidMode(bad) {
			t.Errorf("out-of-range mode %d reported valid", bad)
		}
		if _, err := NewDeclaration(bad); !errors.Is(err, ErrInvalidMode) {
			t.Errorf("NewDeclaration(%d): err = %v, want ErrInvalidMode", bad, err)
		}
	}

	// A declaration is immutable: same-mode redeclare is a no-op, different is
	// refused.
	for _, m := range AllModes {
		d, err := NewDeclaration(m)
		if err != nil {
			t.Fatalf("NewDeclaration(%v): %v", m, err)
		}
		if d.Mode() != m {
			t.Errorf("declared mode = %v, want %v", d.Mode(), m)
		}
		if err := d.CheckImmutable(m); err != nil {
			t.Errorf("redeclaring the same mode %v was refused: %v", m, err)
		}
		for _, other := range AllModes {
			if other == m {
				continue
			}
			if err := d.CheckImmutable(other); !errors.Is(err, ErrModeImmutable) {
				t.Errorf("declaring a different mode %v over %v: err = %v, want ErrModeImmutable", other, m, err)
			}
		}
	}
}
