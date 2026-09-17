// Deterministic Ed25519 signing (T-0103, NFR-006, CQ-004 site #3). Sign is a
// thin wrapper over the standard library crypto/ed25519.Sign: RFC 8032
// Ed25519 signing is deterministic given (key, message), so the only
// non-deterministic input at signing time is the private key material
// itself -- there is no seed, nonce or rand.Reader parameter. Signing the
// same signed-object twice with the same key therefore yields byte-identical
// sig-value (NFR-006).
package eddsa

import "crypto/ed25519"

// SigValueSize is the fixed sig-value length: R (32) || S (32) = 64 octets,
// the RFC 8032 Ed25519 signature encoding (integrity.abnf S5 sig-value).
const SigValueSize = 64

// Sign returns the 64-octet R||S RFC 8032 signature over the 32-octet msg
// (the signed-object digest) under priv. It takes no additional randomness
// parameter (CQ-004 site #3): Ed25519 signing is deterministic, so the
// output depends only on (priv, msg). It delegates entirely to the
// unmodified stdlib crypto/ed25519.Sign.
func Sign(priv ed25519.PrivateKey, msg [32]byte) [SigValueSize]byte {
	sig := ed25519.Sign(priv, msg[:])
	var out [SigValueSize]byte
	copy(out[:], sig)
	return out
}
