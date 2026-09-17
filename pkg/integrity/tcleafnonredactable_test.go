package integrity

import (
	"crypto/sha256"
	"testing"
)

// TestFR_003_TCLeafNonredactableDomainSeparatedFromRedactable is T-0133's
// named test. TCLeafNonredactable returns SHA-256(0x07 || frame), and this
// never collides with the 0x02-tagged redactable leaf of the identical
// frame under any salt (domain separation holds because the domain tags
// 0x07 and 0x02 differ in the first preimage octet).
func TestFR_003_TCLeafNonredactableDomainSeparatedFromRedactable(t *testing.T) {
	frame := []byte{0x02, 0x00, 0xAB, 0xCD, 0xEF}

	got := TCLeafNonredactable(frame)

	// Independently compute SHA-256(0x07 || frame).
	h := sha256.New()
	h.Write([]byte{0x07})
	h.Write(frame)
	var want Digest
	copy(want[:], h.Sum(nil))
	if got != want {
		t.Fatalf("TCLeafNonredactable digest mismatch:\n got  %x\n want %x", got, want)
	}

	// Domain separation: for a sweep of salts, the redactable leaf of the
	// same frame never equals the non-redactable leaf.
	for s := 0; s < 512; s++ {
		var salt [SaltSize]byte
		salt[0] = byte(s)
		salt[1] = byte(s >> 8)
		if TCLeafRedactable(salt, frame) == got {
			t.Fatalf("redactable leaf (salt seed %d) collided with the non-redactable leaf of the same frame", s)
		}
	}

	// A salt of all-zero (the degenerate case) also does not collide.
	var zeroSalt [SaltSize]byte
	if TCLeafRedactable(zeroSalt, frame) == got {
		t.Fatalf("redactable leaf with zero salt collided with the non-redactable leaf")
	}

	// Changing a frame octet changes the non-redactable digest.
	frame2 := append([]byte(nil), frame...)
	frame2[2] ^= 0xFF
	if TCLeafNonredactable(frame2) == got {
		t.Fatalf("changing a frame octet did not change the non-redactable-leaf digest")
	}
}
