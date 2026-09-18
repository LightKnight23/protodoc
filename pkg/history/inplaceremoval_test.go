package history

import (
	"errors"
	"strings"
	"testing"
)

// TestCON_023_RefusesInPlaceRemovalOnCompleteHistory is T-0216's named unit
// test (CON-023). IF an in-place content removal is attempted on a document
// declaring complete history, a conforming implementation refuses the removal
// naming the declared history mode.
func TestCON_023_RefusesInPlaceRemovalOnCompleteHistory(t *testing.T) {
	// Complete history: in-place removal refused, naming the mode.
	complete, _ := NewDeclaration(ModeComplete)
	err := complete.CheckInPlaceRemoval()
	if !errors.Is(err, ErrInPlaceRemovalRefused) {
		t.Fatalf("complete-history in-place removal: err = %v, want ErrInPlaceRemovalRefused", err)
	}
	var ipe *InPlaceRemovalError
	if !errors.As(err, &ipe) {
		t.Fatalf("expected *InPlaceRemovalError, got %T", err)
	}
	if ipe.Mode != ModeComplete {
		t.Errorf("error names mode %v, want complete-history", ipe.Mode)
	}
	// The error message names the declared mode.
	if !strings.Contains(err.Error(), "complete-history") {
		t.Errorf("error message does not name the declared mode: %q", err.Error())
	}

	// Retained-from-point and no-history permit the operation here (their own
	// retention rules apply elsewhere); only complete history forbids it.
	rfp, _ := NewDeclarationWithRetentionPoint(ModeRetainedFromPoint, 1)
	if err := rfp.CheckInPlaceRemoval(); err != nil {
		t.Errorf("retained-from-point in-place removal refused: %v", err)
	}
	none, _ := NewDeclaration(ModeNone)
	if err := none.CheckInPlaceRemoval(); err != nil {
		t.Errorf("no-history in-place removal refused: %v", err)
	}
}
