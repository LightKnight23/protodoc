package merge

import (
	"errors"
	"testing"
)

// TestFR_095_MissingCausalPredecessorBufferedOrRefused is T-0234's named unit
// test (FR-095). IF a transmitted change's causal predecessor is not held by
// the reader, the reader buffers or refuses the change (never applies it) and
// reports the missing predecessor's identifier.
func TestFR_095_MissingCausalPredecessorBufferedOrRefused(t *testing.T) {
	pred := [32]byte{0x11, 0x22}
	held := map[[32]byte]bool{}

	// Missing predecessor, buffer policy -> Buffered + reports the id.
	disp, err := CheckCausalPredecessor(pred, held, true /*buffer*/)
	if disp != Buffered {
		t.Errorf("buffer policy: disposition = %v, want Buffered", disp)
	}
	if disp == Applicable {
		t.Fatal("a change with a missing predecessor must never be Applicable")
	}
	if !errors.Is(err, ErrMissingPredecessor) {
		t.Fatalf("err = %v, want ErrMissingPredecessor", err)
	}
	var me *MissingPredecessorError
	if !errors.As(err, &me) || me.Predecessor != pred {
		t.Fatalf("error must name the missing predecessor %x, got %v", pred, err)
	}

	// Missing predecessor, refuse policy -> Refused + reports the id.
	disp, err = CheckCausalPredecessor(pred, held, false /*refuse*/)
	if disp != Refused {
		t.Errorf("refuse policy: disposition = %v, want Refused", disp)
	}
	if !errors.Is(err, ErrMissingPredecessor) {
		t.Fatalf("err = %v, want ErrMissingPredecessor", err)
	}

	// Predecessor held -> Applicable, no error.
	held[pred] = true
	disp, err = CheckCausalPredecessor(pred, held, false)
	if disp != Applicable || err != nil {
		t.Errorf("held predecessor: disposition=%v err=%v, want Applicable/nil", disp, err)
	}

	// Root change (zero predecessor) -> Applicable, no error.
	disp, err = CheckCausalPredecessor([32]byte{}, map[[32]byte]bool{}, false)
	if disp != Applicable || err != nil {
		t.Errorf("root change: disposition=%v err=%v, want Applicable/nil", disp, err)
	}
}
