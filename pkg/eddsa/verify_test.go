package eddsa

import (
	"crypto/ed25519"
	"testing"
)

// TestCON_015_Verify7StepShortCircuit is T-0108's named test. Verify returns
// true for a genuine signature and false for a bit-flipped one; and an
// instrumented stdlib-call counter proves stdlib Verify is never invoked
// when any of Steps 1-5 already rejected the input.
func TestCON_015_Verify7StepShortCircuit(t *testing.T) {
	_, priv := deterministicKey()
	pub := priv.Public().(ed25519.PublicKey)
	var A [32]byte
	copy(A[:], pub)

	var msg [32]byte
	for i := range msg {
		msg[i] = byte(i * 7)
	}
	sig := Sign(priv, msg)

	// Genuine signature verifies true.
	if !Verify(A, msg, sig) {
		t.Fatalf("genuine signature failed to verify")
	}

	// Bit-flipped signature verifies false.
	bad := sig
	bad[10] ^= 0x01
	if Verify(A, msg, bad) {
		t.Fatalf("bit-flipped signature verified true")
	}

	// Wrong message verifies false.
	var msg2 [32]byte
	copy(msg2[:], msg[:])
	msg2[0] ^= 0xFF
	if Verify(A, msg2, sig) {
		t.Fatalf("signature verified against the wrong message")
	}

	// --- Short-circuit: stdlib Verify is never called when Steps 1-5
	// reject. Install a counting stub. ---
	orig := stdlibVerify
	t.Cleanup(func() { stdlibVerify = orig })
	calls := 0
	stdlibVerify = func(pub ed25519.PublicKey, message, s []byte) bool {
		calls++
		return orig(pub, message, s)
	}

	// A rejected at Step 2 (small-order A): stdlib must not be called.
	calls = 0
	if Verify(smallOrderPoints[0], msg, sig) {
		t.Fatalf("small-order A verified true")
	}
	if calls != 0 {
		t.Fatalf("stdlib Verify called %d times on a Step-2 rejection, want 0", calls)
	}

	// R rejected at Step 4 (small-order R): stdlib must not be called.
	calls = 0
	var soR [SigValueSize]byte
	copy(soR[:32], smallOrderPoints[4][:]) // small-order R
	copy(soR[32:], sig[32:])
	if Verify(A, msg, soR) {
		t.Fatalf("small-order R verified true")
	}
	if calls != 0 {
		t.Fatalf("stdlib Verify called %d times on a Step-4 rejection, want 0", calls)
	}

	// S out of range (Step 5): set S = all 0xFF (>= L). stdlib must not be called.
	calls = 0
	var badS [SigValueSize]byte
	copy(badS[:32], sig[:32])
	for i := 32; i < 64; i++ {
		badS[i] = 0xFF
	}
	if Verify(A, msg, badS) {
		t.Fatalf("out-of-range S verified true")
	}
	if calls != 0 {
		t.Fatalf("stdlib Verify called %d times on a Step-5 rejection, want 0", calls)
	}

	// A genuine input DOES reach stdlib (exactly once).
	calls = 0
	if !Verify(A, msg, sig) {
		t.Fatalf("genuine signature failed after stub install")
	}
	if calls != 1 {
		t.Fatalf("stdlib Verify called %d times on a valid input, want exactly 1", calls)
	}
}
