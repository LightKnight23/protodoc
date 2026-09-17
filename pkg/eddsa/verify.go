// EdDSA-Protodoc-1 Verify (T-0108, CON-015): the closed 7-step
// verification procedure (integrity.abnf S6). Steps 1-5 are the
// curve-arithmetic-free pre-filters (checkPublicKey, checkR, checkScalarS);
// on the first failure Verify returns false immediately WITHOUT calling
// stdlib, so no later step's cost or behaviour is observable on a rejected
// input. When all 5 pass, Steps 6-7 are delegated verbatim to the
// unmodified stdlib crypto/ed25519.Verify(A, msg, R||S); its boolean result
// is the procedure's result. The Step 6 challenge k is never recomputed
// outside that stdlib call.
package eddsa

import "crypto/ed25519"

// stdlibVerify is the delegation target for Steps 6-7. It is a package var
// solely so a test can substitute a call-counting stub to prove the
// short-circuit never reaches stdlib on a Step-1..5 rejection; production
// code never reassigns it, and it is exactly crypto/ed25519.Verify.
var stdlibVerify = func(pub ed25519.PublicKey, message, sig []byte) bool {
	return ed25519.Verify(pub, message, sig)
}

// Verify runs the full EdDSA-Protodoc-1 procedure. A is the 32-octet public
// key; msg is the 32-octet signed-object digest; sig is the 64-octet R||S
// sig-value. It returns true iff Steps 1-5 all pass AND unmodified stdlib
// crypto/ed25519.Verify accepts (A, msg, R||S).
//
// Steps 1-5 short-circuit: the first failing check returns false and stdlib
// is never invoked, closing the timing/behavioural side channel between an
// early-step rejection and a Step-7 rejection.
func Verify(A [32]byte, msg [32]byte, sig [SigValueSize]byte) bool {
	var R, S [32]byte
	copy(R[:], sig[:32])
	copy(S[:], sig[32:])

	// Step 1-2: public key A canonical + non-small-order.
	if checkPublicKey(A) != nil {
		return false
	}
	// Step 3-4: R canonical + non-small-order.
	if checkR(R) != nil {
		return false
	}
	// Step 5: scalar S in range.
	if checkScalarS(S) != nil {
		return false
	}
	// Step 6-7: delegate to unmodified stdlib; its result is the verdict.
	return stdlibVerify(ed25519.PublicKey(A[:]), msg[:], sig[:])
}
