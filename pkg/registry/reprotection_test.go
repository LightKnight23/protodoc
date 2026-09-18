package registry_test

import (
	"crypto/ed25519"
	"testing"

	"Protodoc/pkg/eddsa"
	"Protodoc/pkg/registry"
)

// TestCON_016_ReProtectionPreservesOriginalVerdict is T-0303's named
// integration test (CON-016, hash-family-upgrade half). A document re-protected
// under SHA3-256 still verifies its original Ed25519 signature to the SAME
// verdict as before; the re-protection-aware verifier additionally checks the
// SHA3-256 outer digest; applying re-protection twice does not corrupt the
// original signature.
func TestCON_016_ReProtectionPreservesOriginalVerdict(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	var signedObject [32]byte
	for i := range signedObject {
		signedObject[i] = byte(i * 7)
	}

	// Original signature over the signed_object.
	sig := eddsa.Sign(priv, signedObject)

	verify := func() bool {
		return ed25519.Verify(pub, signedObject[:], sig[:])
	}

	// Verdict BEFORE re-protection: valid.
	if !verify() {
		t.Fatalf("original signature must verify before re-protection")
	}

	// Re-protect under SHA3-256 with a fresh time attestation.
	rp := registry.ReProtect(signedObject[:], []byte("time-attestation-1"))

	// Original signature verdict UNCHANGED after re-protection.
	if !verify() {
		t.Errorf("original signature verdict changed after re-protection; must be intact")
	}
	// The outer SHA3-256 layer verifies.
	if !registry.VerifyReProtection(rp, signedObject[:]) {
		t.Errorf("SHA3-256 outer digest must verify against the signed_object")
	}

	// Re-protect a SECOND time: original signature still intact, and the outer
	// digest over the same signed_object is unchanged (idempotent inner wrap).
	rp2 := registry.ReProtect(signedObject[:], []byte("time-attestation-2"))
	if !verify() {
		t.Errorf("original signature must survive a second re-protection")
	}
	if rp2.OuterDigest != rp.OuterDigest {
		t.Errorf("outer digest over the same signed_object must be stable across re-protections")
	}
	if !registry.VerifyReProtection(rp2, signedObject[:]) {
		t.Errorf("second re-protection's outer digest must verify")
	}

	// A tampered signed_object fails the outer check (proving it binds bytes).
	tampered := append([]byte(nil), signedObject[:]...)
	tampered[0] ^= 0xFF
	if registry.VerifyReProtection(rp, tampered) {
		t.Errorf("outer digest must not verify against tampered signed_object bytes")
	}
}
