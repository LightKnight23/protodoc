package history

import (
	"crypto/sha256"
	"testing"

	"Protodoc/pkg/integrity"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_061_ErasureRecordSaltedCommitmentForm is T-0217's named unit test
// (FR-061). An ErasureRecord enumerates the severed state's identity and a
// SALTED-commitment digest SHA-256(0x09 || salt || digest); the salt is
// destroyed at trim, leaving a hiding commitment. This asserts the commitment
// form, the salt-destruction, the byte-exact round trip, and the hiding
// property (the commitment binds the digest under a hidden salt).
func TestFR_061_ErasureRecordSaltedCommitmentForm(t *testing.T) {
	identity := hUnitID(0x01)
	var digest pdlfmt.Digest256
	for i := range digest {
		digest[i] = byte(0x40 + i)
	}
	var salt [SaltSize]byte
	for i := range salt {
		salt[i] = byte(0x90 + i)
	}

	er := NewErasureRecord(identity, digest, salt)

	// The commitment is exactly SHA-256(0x09 || salt || digest).
	h := sha256.New()
	h.Write([]byte{integrity.DomainSeveranceCommitment})
	h.Write(salt[:])
	h.Write(digest[:])
	var want pdlfmt.Digest256
	copy(want[:], h.Sum(nil))
	if er.Commitment != want {
		t.Error("commitment is not SHA-256(0x09 || salt || digest)")
	}

	// Byte-exact round trip (with salt present, the brief pre-trim window).
	enc, err := er.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	dec, err := DecodeErasureRecord(enc)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if dec != er {
		t.Error("erasure record did not survive round trip")
	}

	// Trim destroys the salt; the identity, digest, and commitment remain.
	trimmed := er.Trimmed()
	if !trimmed.SaltDestroyed() {
		t.Error("salt not destroyed after trim")
	}
	if trimmed.Identity != identity || trimmed.Digest != digest || trimmed.Commitment != er.Commitment {
		t.Error("trim altered identity/digest/commitment")
	}
	// The trimmed record round-trips too (salt all-zero).
	tEnc, _ := trimmed.Encode()
	tDec, err := DecodeErasureRecord(tEnc)
	if err != nil || tDec != trimmed {
		t.Errorf("trimmed record round trip failed: %v", err)
	}

	// Hiding: a different salt over the same digest yields a different
	// commitment (so a destroyed salt hides the digest).
	var salt2 [SaltSize]byte
	salt2[0] = 0xFF
	if SeveranceCommitment(salt2, digest) == er.Commitment {
		t.Error("two salts over the same digest produced the same commitment (not hiding)")
	}
	// Binding: a different digest under the same salt yields a different
	// commitment.
	var digest2 pdlfmt.Digest256
	digest2[0] = 0xFF
	if SeveranceCommitment(salt, digest2) == er.Commitment {
		t.Error("two digests under the same salt produced the same commitment (not binding)")
	}

	// The severance-commitment domain tag is the shared 0x09 (reused, not a
	// second definition).
	if integrity.DomainSeveranceCommitment != 0x09 {
		t.Errorf("DomainSeveranceCommitment = 0x%02x, want 0x09", integrity.DomainSeveranceCommitment)
	}
}
