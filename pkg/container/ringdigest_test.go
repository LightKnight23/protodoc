package container

import "testing"

// TestFR_117_SelfDigestDetectsTornSlot is T-0009's named unit test:
// a slot with a corrupted single octet fails self-digest validation, an
// untouched slot passes, and validation requires no read outside the
// slot's own bytes. Implements: FR-117.
func TestFR_117_SelfDigestDetectsTornSlot(t *testing.T) {
	rec := fixtureRingRecord(42, StateID{})
	slot := rec.Encode(nil)
	if len(slot) != RingSlotSize {
		t.Fatalf("fixture slot length = %d, want %d (no read outside the slot's own bytes is possible otherwise)", len(slot), RingSlotSize)
	}

	if !VerifyRingSlotDigest(slot) {
		t.Fatal("untouched slot failed self-digest validation, want pass")
	}

	// Corrupt a single octet inside ring-fields, well within [0,480).
	corruptedField := append([]byte(nil), slot...)
	corruptedField[10] ^= 0xFF
	if VerifyRingSlotDigest(corruptedField) {
		t.Fatal("slot with one corrupted ring-fields octet passed self-digest validation, want fail")
	}

	// Corrupting the stored record-digest itself, not the fields it
	// covers, must also be detected.
	corruptedDigest := append([]byte(nil), slot...)
	corruptedDigest[RingSlotSize-1] ^= 0xFF
	if VerifyRingSlotDigest(corruptedDigest) {
		t.Fatal("slot with corrupted record-digest passed self-digest validation, want fail")
	}
}

func TestVerifyRingSlotDigest_RejectsNeverWrittenSlot(t *testing.T) {
	var zero [RingSlotSize]byte
	if VerifyRingSlotDigest(zero[:]) {
		t.Fatal("all-zero (never-written) slot passed self-digest validation, want fail")
	}
}

func TestVerifyRingSlotDigest_RejectsTruncatedSlot(t *testing.T) {
	if VerifyRingSlotDigest(make([]byte, RingSlotSize-1)) {
		t.Fatal("truncated slot passed self-digest validation, want fail")
	}
}
