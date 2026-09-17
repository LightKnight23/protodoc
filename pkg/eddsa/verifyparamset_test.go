package eddsa

import (
	"crypto/ed25519"
	"testing"
)

// TestCON_015_OffAllowlistParamSetRejected is T-0111's named test. A
// call-counting stub proves Verify is never invoked when the parameter set
// fails Validate; a table enumerates 0x0000, 0x0002 and 0xFFFF as rejected
// (with Verify never reached) and 0x0001 as the only value that proceeds to
// Verify.
func TestCON_015_OffAllowlistParamSetRejected(t *testing.T) {
	_, priv := deterministicKey()
	pub := priv.Public().(ed25519.PublicKey)
	var A [32]byte
	copy(A[:], pub)
	var msg [32]byte
	msg[0] = 0x5A
	sig := Sign(priv, msg)

	// Install a counting stub for the verification target.
	orig := verifyFunc
	t.Cleanup(func() { verifyFunc = orig })
	calls := 0
	verifyFunc = func(a [32]byte, m [32]byte, s [SigValueSize]byte) bool {
		calls++
		return orig(a, m, s)
	}

	cases := []struct {
		name        string
		paramSet    ParamSet
		wantProceed bool // whether Verify should be reached
	}{
		{"reserved-0x0000", ParamSetReserved, false},
		{"reserved-0x0002", 0x0002, false},
		{"reserved-0xFFFF", 0xFFFF, false},
		{"allowlisted-0x0001", ParamSetV1, true},
	}
	for _, c := range cases {
		calls = 0
		got := VerifyWithParamSet(c.paramSet, A, msg, sig)
		if !c.wantProceed {
			// Off-allowlist: rejected (false) and Verify never called.
			if got {
				t.Fatalf("%s: VerifyWithParamSet returned true, want rejected", c.name)
			}
			if calls != 0 {
				t.Fatalf("%s: Verify called %d times, want 0 (must short-circuit before the arithmetic)", c.name, calls)
			}
		} else {
			// Allowlisted: proceeds to Verify (called exactly once) and, on
			// a genuine signature, returns true.
			if calls != 1 {
				t.Fatalf("%s: Verify called %d times, want exactly 1", c.name, calls)
			}
			if !got {
				t.Fatalf("%s: genuine signature under the allowlisted param_set failed", c.name)
			}
		}
	}
}
