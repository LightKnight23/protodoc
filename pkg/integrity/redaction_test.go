package integrity

import (
	"errors"
	"testing"
)

func redSalt(seed byte) [SaltSize]byte {
	var s [SaltSize]byte
	for i := range s {
		s[i] = seed + byte(i)
	}
	return s
}

func redUnitID(seed byte) UnitID {
	var id UnitID
	for i := range id {
		id[i] = seed + byte(i)
	}
	return id
}

// TestFR_074_DesignateRedactableSubtree is T-0185's named unit test (FR-074).
// WHEN a document is signed, the signatory may designate content subtrees as
// redactable, each committed with a hiding-and-binding commitment. A
// designated subtree reports designated, produces its salted commitment, and
// carries its salt inside it; a subtree not designated produces no commitment.
func TestFR_074_DesignateRedactableSubtree(t *testing.T) {
	frame := []byte("a redactable text block's stored frame octets")
	salt := redSalt(0x11)
	sub := DesignateRedactable(redUnitID(0x01), frame, salt)

	if !sub.IsDesignated() {
		t.Error("designated subtree reports not designated")
	}
	if !sub.SaltStoredInside() {
		t.Error("salt should be stored inside the subtree while present")
	}
	if !sub.SaltPresent() {
		t.Error("a non-zero salt should be present before removal")
	}

	commit, err := sub.Commitment()
	if err != nil {
		t.Fatalf("Commitment: %v", err)
	}
	// The commitment is exactly the salted redactable-leaf digest.
	want := TCLeafRedactable(salt, frame)
	if commit != want {
		t.Errorf("commitment != SHA-256(0x02 || salt || frame)")
	}

	// A subtree NOT designated redactable produces no commitment.
	notDesignated := RedactableSubtree{UnitID: redUnitID(0x02), Frame: frame, Salt: salt}
	if notDesignated.IsDesignated() {
		t.Error("undesignated subtree reports designated")
	}
	if _, err := notDesignated.Commitment(); !errors.Is(err, ErrNotDesignated) {
		t.Errorf("undesignated Commitment: err = %v, want ErrNotDesignated", err)
	}

	// Designation copies the frame (caller mutation does not affect it).
	frame[0] = 'X'
	commit2, _ := sub.Commitment()
	if commit2 != want {
		t.Error("mutating the caller's frame slice changed the subtree commitment (frame not copied)")
	}
}
