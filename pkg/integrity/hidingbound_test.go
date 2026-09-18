package integrity

import "testing"

// TestFR_075_HidingBoundAgainstBruteForce is T-0189's named unit test
// (FR-075). After redaction the salt is erased, so the retained value is the
// bare commitment SHA-256(0x02 || salt || frame) with the 32-octet salt gone.
// The omitted content is therefore not recoverable from the retained integrity
// values by exhaustive search over any candidate set below 2^80: even an
// attacker who KNOWS the exact candidate plaintext cannot confirm it against
// the commitment without also finding the 256-bit salt, whose search space
// (2^256) exceeds the 2^80 floor by 176 bits.
func TestFR_075_HidingBoundAgainstBruteForce(t *testing.T) {
	secret := []byte("the redacted secret paragraph")
	salt := redSalt(0x7F)
	sub := DesignateRedactable(redUnitID(1), secret, salt)
	commitment, _ := sub.Commitment()

	// Redact: the salt is now gone from the file; only `commitment` remains.
	if err := sub.Remove(); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if sub.SaltPresent() {
		t.Fatal("salt must be gone after removal for the hiding bound to hold")
	}

	// An attacker who knows the exact candidate plaintext still cannot confirm
	// it against the retained commitment without the salt: recomputing with any
	// salt OTHER than the (now-erased) original does not match. We model the
	// attacker's bounded search as trying a modest candidate-salt set; none of
	// them reproduces the commitment (the real salt is not in the attacker's
	// small search space, and 2^256 >> 2^80 makes finding it infeasible).
	knownPlaintext := secret
	matches := 0
	for guessByte := 0; guessByte < 4096; guessByte++ {
		var guessSalt [SaltSize]byte
		guessSalt[0] = byte(guessByte)
		guessSalt[1] = byte(guessByte >> 8)
		if TCLeafRedactable(guessSalt, knownPlaintext) == commitment {
			matches++
		}
	}
	if matches != 0 {
		t.Errorf("a bounded salt search recovered the commitment for known plaintext (%d matches); the hiding bound is broken", matches)
	}

	// The hiding floor is documented as 2^80; the actual post-removal search
	// (finding the 256-bit salt) is 2^256, exceeding the floor. Assert the
	// salt width backing this is the full 32 octets.
	if SaltSize*8 < 80 {
		t.Errorf("salt width %d bits is below the FR-075 hiding floor of 80 bits", SaltSize*8)
	}
	if SaltSize != 32 {
		t.Errorf("salt width = %d octets, want 32 (256-bit search space)", SaltSize)
	}

	// Binding sanity: the retained commitment still binds the original content
	// under the original salt (a verifier with the salt could confirm; without
	// it, cannot). Recomputing with the true salt matches.
	if TCLeafRedactable(salt, secret) != commitment {
		t.Error("the retained commitment does not bind the original (salt, content)")
	}
}
