// Off-allowlist parameter-set rejection gate (T-0111, CON-015).
// VerifyWithParamSet validates the parameter set BEFORE any cryptographic
// verification: an off-allowlist param_set (0x0000, or any reserved
// 0x0002-0xFFFF) is rejected outright -- never a positive or warning verdict
// -- short-circuiting before the Ed25519 arithmetic (Verify) is ever
// reached, so an off-allowlist selection can never touch the curve.
package eddsa

// verifyFunc is the verification target VerifyWithParamSet calls once the
// parameter set is validated. It is a package var only so a test can
// substitute a call-counting stub to prove the gate short-circuits before
// verification; production code never reassigns it, and it is exactly Verify.
var verifyFunc = Verify

// VerifyWithParamSet verifies a signature only under an allowlisted
// parameter set. It calls paramSet.Validate() first: on failure it returns
// false immediately (the off-allowlist rejection), without invoking Verify,
// so the Ed25519 procedure is never reached for a non-allowlisted param_set.
// On the sole allowlisted value (ParamSetV1) it delegates to the full
// 7-step Verify and returns its result.
func VerifyWithParamSet(paramSet ParamSet, A [32]byte, msg [32]byte, sig [SigValueSize]byte) bool {
	if paramSet.Validate() != nil {
		return false
	}
	return verifyFunc(A, msg, sig)
}
