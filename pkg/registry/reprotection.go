// Re-protection (CON-016, hash-family-upgrade half; T-0303). Re-protection
// wraps an existing signed_object with a FRESH OUTER DIGEST computed under a
// stronger/alternate hash family (SHA3-256) plus a fresh time attestation,
// WITHOUT re-signing. The original Ed25519 signature and its verdict remain
// intact and independently re-checkable; the re-protection-aware verifier
// additionally checks the new outer digest. Applying re-protection twice does
// not corrupt or invalidate the original signature (the outer layer wraps, it
// never rewrites the inner signed_object).
package registry

import "crypto/sha3"

// ReProtection is the outer protection layer added over an unmodified
// signed_object: a SHA3-256 digest over the inner signed_object bytes plus a
// fresh time-attestation token. It carries NO new Ed25519 signature.
type ReProtection struct {
	// OuterDigest is SHA3-256 over the wrapped signed_object bytes.
	OuterDigest [32]byte
	// TimeAttestation is the fresh time-attestation token attached at
	// re-protection (opaque here; its shape is integrity.abnf's concern).
	TimeAttestation []byte
}

// ReProtect wraps signedObject with a fresh SHA3-256 outer digest and the given
// time attestation, returning the outer layer. It does NOT modify signedObject
// and produces no new signature. Re-protecting an already-re-protected document
// wraps the same inner signed_object again (idempotent over the inner bytes):
// the outer digest is a pure function of the inner signed_object, so a second
// application over the same signed_object yields the same outer digest.
func ReProtect(signedObject []byte, timeAttestation []byte) ReProtection {
	ta := make([]byte, len(timeAttestation))
	copy(ta, timeAttestation)
	return ReProtection{
		OuterDigest:     sha3.Sum256(signedObject),
		TimeAttestation: ta,
	}
}

// VerifyReProtection checks the outer SHA3-256 digest against a fresh recompute
// over the (unmodified) signed_object bytes. It returns true when the outer
// layer is intact. It does NOT touch the inner Ed25519 signature — that is
// verified independently and unchanged.
func VerifyReProtection(rp ReProtection, signedObject []byte) bool {
	return sha3.Sum256(signedObject) == rp.OuterDigest
}
