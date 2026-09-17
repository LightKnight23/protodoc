package eddsa

import (
	"crypto/ed25519"
	"crypto/sha256"
	"testing"
)

// TestNFR_006_SignTwiceIdenticalOctets is T-0110's named test: the literal
// NFR-006 acceptance test. For representative messages of 0, 1 and 4096
// octets, it hashes each to a 32-octet signed-object, signs it twice with
// the same key, and asserts the raw 64-octet sig-value is byte-identical
// both times, then confirms both outputs verify true.
func TestNFR_006_SignTwiceIdenticalOctets(t *testing.T) {
	_, priv := deterministicKey()
	pub := priv.Public().(ed25519.PublicKey)
	var A [32]byte
	copy(A[:], pub)

	sizes := []int{0, 1, 4096}
	for _, n := range sizes {
		message := make([]byte, n)
		for i := range message {
			message[i] = byte(i % 251)
		}
		signedObject := sha256.Sum256(message) // 32-octet digest that is signed

		sig1 := Sign(priv, signedObject)
		sig2 := Sign(priv, signedObject)

		if sig1 != sig2 {
			t.Fatalf("message size %d: two Sign() calls produced different sig-value (non-deterministic)", n)
		}
		if !Verify(A, signedObject, sig1) {
			t.Fatalf("message size %d: sig1 failed Verify", n)
		}
		if !Verify(A, signedObject, sig2) {
			t.Fatalf("message size %d: sig2 failed Verify", n)
		}
	}
}
