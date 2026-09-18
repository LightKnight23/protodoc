package integrity

import "testing"

// TestFR_074_SaltedCommitmentDigest is T-0187's named unit test (FR-074/FR-075).
// The redactable-subtree commitment is a hiding-and-binding commitment:
//
//   - HIDING: the same content under two different salts yields two different
//     commitments, so the retained digest reveals nothing about the content
//     without the salt (the basis of FR-075's 2^80 hiding floor).
//   - BINDING: two different contents under the same salt yield two different
//     commitments, so a signatory cannot later claim a different content for a
//     given commitment.
//   - DETERMINISTIC: the same (salt, content) always yields the same
//     commitment (a verifier recomputes it).
func TestFR_074_SaltedCommitmentDigest(t *testing.T) {
	frameA := []byte("content A")
	frameB := []byte("content B")
	salt1 := redSalt(0x10)
	salt2 := redSalt(0x20)

	c := func(salt [SaltSize]byte, frame []byte) Digest {
		d, err := DesignateRedactable(redUnitID(1), frame, salt).Commitment()
		if err != nil {
			t.Fatalf("Commitment: %v", err)
		}
		return d
	}

	// Deterministic: same salt + same content -> same commitment.
	if c(salt1, frameA) != c(salt1, frameA) {
		t.Error("commitment is not deterministic for the same (salt, content)")
	}

	// Hiding: same content, different salts -> different commitments.
	if c(salt1, frameA) == c(salt2, frameA) {
		t.Error("same content under two salts produced the same commitment (not hiding)")
	}

	// Binding: different content, same salt -> different commitments.
	if c(salt1, frameA) == c(salt1, frameB) {
		t.Error("two contents under the same salt produced the same commitment (not binding)")
	}

	// Domain separation from the non-redactable leaf: no salt can make a
	// redactable commitment collide with the non-redactable leaf of the same
	// frame (different domain tags 0x02 vs 0x07).
	if c(salt1, frameA) == TCLeafNonredactable(frameA) {
		t.Error("redactable commitment collided with the non-redactable leaf of the same frame")
	}
}
