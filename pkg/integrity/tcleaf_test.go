package integrity

import (
	"crypto/sha256"
	"testing"
)

// TestFR_003_TCLeafRedactableUsesStoredFrameVerbatim is T-0132's named test.
// TCLeafRedactable returns SHA-256(0x02 || salt || frame) with frame taken
// byte-for-byte (no re-encode step), and changing one octet of the stored
// frame -- or of the salt -- changes the leaf digest.
func TestFR_003_TCLeafRedactableUsesStoredFrameVerbatim(t *testing.T) {
	var salt [SaltSize]byte
	for i := range salt {
		salt[i] = byte(i + 1)
	}
	frame := []byte{0x01, 0x00, 0x11, 0x22, 0x33, 0x44} // a stored record frame

	got := TCLeafRedactable(salt, frame)

	// Independently compute the expected digest: SHA-256(0x02 || salt || frame).
	h := sha256.New()
	h.Write([]byte{0x02})
	h.Write(salt[:])
	h.Write(frame)
	var want Digest
	copy(want[:], h.Sum(nil))
	if got != want {
		t.Fatalf("TCLeafRedactable digest mismatch:\n got  %x\n want %x", got, want)
	}

	// Changing one octet of the frame changes the digest.
	frame2 := append([]byte(nil), frame...)
	frame2[3] ^= 0xFF
	if TCLeafRedactable(salt, frame2) == got {
		t.Fatalf("changing a frame octet did not change the redactable-leaf digest")
	}

	// Changing one octet of the salt changes the digest.
	salt2 := salt
	salt2[10] ^= 0xFF
	if TCLeafRedactable(salt2, frame) == got {
		t.Fatalf("changing a salt octet did not change the redactable-leaf digest")
	}

	// The frame is used verbatim: an empty frame with the same salt differs
	// from a non-empty frame.
	if TCLeafRedactable(salt, nil) == got {
		t.Fatalf("empty frame produced the same digest as a non-empty frame")
	}
}
