package history

import (
	"errors"
	"testing"
)

// TestCON_022_RetentionPointModeGated is T-0213's named unit test (CON-022).
// A retention point is meaningful ONLY when history-mode is
// retained-from-point; declaring a retention point under any other mode is
// refused, and a retained-from-point declaration carries a meaningful point.
func TestCON_022_RetentionPointModeGated(t *testing.T) {
	// retained-from-point: a retention point is accepted and meaningful.
	d, err := NewDeclarationWithRetentionPoint(ModeRetainedFromPoint, 42)
	if err != nil {
		t.Fatalf("retained-from-point with retention point: %v", err)
	}
	if d.RetentionPoint() != 42 {
		t.Errorf("retention point = %d, want 42", d.RetentionPoint())
	}
	if !d.HasMeaningfulRetentionPoint() {
		t.Error("retained-from-point declaration should have a meaningful retention point")
	}

	// complete / no-history: a retention point is mode-gated (refused).
	for _, m := range []Mode{ModeComplete, ModeNone} {
		if _, err := NewDeclarationWithRetentionPoint(m, 42); !errors.Is(err, ErrRetentionPointModeGated) {
			t.Errorf("retention point under mode %v: err = %v, want ErrRetentionPointModeGated", m, err)
		}
		// A plain declaration of those modes has no meaningful retention point.
		plain, _ := NewDeclaration(m)
		if plain.HasMeaningfulRetentionPoint() {
			t.Errorf("mode %v should have no meaningful retention point", m)
		}
		if plain.RetentionPoint() != 0 {
			t.Errorf("mode %v retention point = %d, want 0", m, plain.RetentionPoint())
		}
	}

	// An invalid mode is still rejected.
	if _, err := NewDeclarationWithRetentionPoint(Mode(7), 1); !errors.Is(err, ErrInvalidMode) {
		t.Errorf("invalid mode: err = %v, want ErrInvalidMode", err)
	}
}
