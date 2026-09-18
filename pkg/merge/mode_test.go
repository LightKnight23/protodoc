package merge

import (
	"errors"
	"strings"
	"testing"

	"Protodoc/pkg/history"
)

// TestCON_025_RefuseMergeOnHistoryModeMismatch is T-0237's named unit test
// (CON-025). IF two documents/branches declare different history modes, the
// merge is refused naming BOTH declared modes, with zero partial merge
// outputs. Matching modes are permitted.
func TestCON_025_RefuseMergeOnHistoryModeMismatch(t *testing.T) {
	decl := func(m history.Mode) history.Declaration {
		var d history.Declaration
		var err error
		if m == history.ModeRetainedFromPoint {
			d, err = history.NewDeclarationWithRetentionPoint(m, 10)
		} else {
			d, err = history.NewDeclaration(m)
		}
		if err != nil {
			t.Fatalf("declare %s: %v", history.ModeName(m), err)
		}
		return d
	}

	modes := []history.Mode{history.ModeComplete, history.ModeRetainedFromPoint, history.ModeNone}

	// Every distinct-mode pair is refused, naming both modes; zero outputs.
	partialOutputs := 0
	attempts := 0
	for _, a := range modes {
		for _, b := range modes {
			if a == b {
				continue
			}
			for i := 0; i < 9; i++ { // >= 50 attempts across the 6 ordered pairs
				attempts++
				err := CheckHistoryModeMatch(decl(a), decl(b))
				if !errors.Is(err, ErrHistoryModeMismatch) {
					t.Fatalf("(%s,%s): err = %v, want ErrHistoryModeMismatch",
						history.ModeName(a), history.ModeName(b), err)
				}
				var me *HistoryModeMismatchError
				if !errors.As(err, &me) {
					t.Fatalf("expected *HistoryModeMismatchError, got %T", err)
				}
				// BOTH modes named in the message.
				msg := me.Error()
				if !strings.Contains(msg, history.ModeName(a)) || !strings.Contains(msg, history.ModeName(b)) {
					t.Errorf("message %q must name both modes %s and %s",
						msg, history.ModeName(a), history.ModeName(b))
				}
				if err == nil {
					partialOutputs++
				}
			}
		}
	}
	if attempts < 50 {
		t.Errorf("ran %d cross-mode attempts, want >= 50", attempts)
	}
	if partialOutputs != 0 {
		t.Errorf("cross-mode merges produced %d partial outputs, want 0", partialOutputs)
	}

	// Matching modes are permitted.
	for _, m := range modes {
		if err := CheckHistoryModeMatch(decl(m), decl(m)); err != nil {
			t.Errorf("matching mode %s should merge, got %v", history.ModeName(m), err)
		}
	}
}
