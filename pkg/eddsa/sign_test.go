package eddsa

import (
	"crypto/ed25519"
	"reflect"
	"testing"
)

// deterministicKey returns a fixed Ed25519 key pair from a constant seed so
// the determinism test is itself reproducible (the seed is test-only, not a
// production key source).
func deterministicKey() (ed25519.PublicKey, ed25519.PrivateKey) {
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = byte(i + 1)
	}
	priv := ed25519.NewKeyFromSeed(seed)
	return priv.Public().(ed25519.PublicKey), priv
}

// TestNFR_006_SignDeterministic is T-0103's named test. Sign returns the
// 64-octet R||S encoding, is deterministic (same key+message -> identical
// output), and its signature accepts no seed/nonce/rand.Reader parameter
// beyond the key and message (verified by reflection).
func TestNFR_006_SignDeterministic(t *testing.T) {
	_, priv := deterministicKey()

	var msg [32]byte
	for i := range msg {
		msg[i] = byte(i * 3)
	}

	a := Sign(priv, msg)
	b := Sign(priv, msg)
	if a != b {
		t.Fatalf("Sign is not deterministic: two calls produced different sig-value")
	}
	if len(a) != SigValueSize {
		t.Fatalf("sig-value length %d, want %d", len(a), SigValueSize)
	}

	// A different message yields a different signature (the message is
	// actually signed).
	var msg2 [32]byte
	copy(msg2[:], msg[:])
	msg2[0] ^= 0xFF
	if Sign(priv, msg2) == a {
		t.Fatalf("different messages produced identical sig-value")
	}

	// The signature is a valid RFC 8032 signature accepted by stdlib.
	pub := priv.Public().(ed25519.PublicKey)
	if !ed25519.Verify(pub, msg[:], a[:]) {
		t.Fatalf("Sign output does not verify under stdlib ed25519.Verify")
	}

	// Signature shape: Sign takes exactly (ed25519.PrivateKey, [32]byte) and
	// returns [64]byte -- no rand.Reader/seed/nonce parameter (CQ-004 #3).
	ft := reflect.TypeOf(Sign)
	if ft.NumIn() != 2 {
		t.Fatalf("Sign takes %d parameters, want 2 (key, message) -- no randomness parameter", ft.NumIn())
	}
	if ft.In(0) != reflect.TypeOf(ed25519.PrivateKey(nil)) {
		t.Fatalf("Sign first param is %v, want ed25519.PrivateKey", ft.In(0))
	}
	if ft.In(1).Kind() != reflect.Array || ft.In(1).Len() != 32 {
		t.Fatalf("Sign second param is %v, want [32]byte", ft.In(1))
	}
}
