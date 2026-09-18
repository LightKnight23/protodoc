package integrity

import (
	"bytes"
	"errors"
	"testing"
)

// TestFR_074_SaltStoredInsideSubtree is T-0186's named unit test (FR-074). The
// commitment's salt is stored INSIDE the designated subtree, so the act of
// redaction (Remove) deletes both the subtree's frame octets AND its salt in
// one operation, leaving only the bare 32-octet commitment digest -- no salt,
// no plaintext, anywhere.
func TestFR_074_SaltStoredInsideSubtree(t *testing.T) {
	frame := []byte("secret content to be redacted")
	salt := redSalt(0x22)
	sub := DesignateRedactable(redUnitID(0x01), frame, salt)

	// Capture the commitment while present.
	before, err := sub.Commitment()
	if err != nil {
		t.Fatalf("Commitment: %v", err)
	}

	// The salt is present and stored inside before removal.
	if !sub.SaltStoredInside() || !sub.SaltPresent() {
		t.Fatal("salt should be present and stored inside before removal")
	}

	// Remove: deletes frame AND salt together.
	if err := sub.Remove(); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if !sub.Removed() {
		t.Error("subtree should report removed")
	}

	// The salt is gone (no non-zero salt remains) and the frame is nil.
	if sub.SaltPresent() {
		t.Error("salt should be gone after removal")
	}
	if sub.SaltStoredInside() {
		t.Error("salt should no longer be stored inside after removal")
	}
	if !bytes.Equal(sub.Salt[:], zeroSalt[:]) {
		t.Error("salt octets should be zeroed after removal")
	}
	if sub.Frame != nil {
		t.Error("frame should be nil after removal")
	}

	// The retained bare commitment digest is unchanged (so the signed T_C
	// still verifies).
	after, err := sub.Commitment()
	if err != nil {
		t.Fatalf("Commitment after removal: %v", err)
	}
	if after != before {
		t.Error("commitment digest changed after removal; it must be retained unchanged")
	}

	// A second Remove errors rather than corrupting the retained commitment.
	if err := sub.Remove(); !errors.Is(err, ErrAlreadyRemoved) {
		t.Errorf("second Remove: err = %v, want ErrAlreadyRemoved", err)
	}
	again, _ := sub.Commitment()
	if again != before {
		t.Error("retained commitment corrupted by a second Remove")
	}
}
