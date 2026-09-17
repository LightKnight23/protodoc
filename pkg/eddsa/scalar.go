// EdDSA-Protodoc-1 Step 5 (integrity.abnf S6): the scalar range check on
// the S component of sig-value. S, decoded as a little-endian 256-bit
// unsigned integer, must be strictly less than the group order
// L = 2^252 + 27742317777372353535851937790883648493 (RFC 8032 exact
// value); S >= L is a non-canonical scalar and is rejected. This closes the
// scalar-malleability divergence (implementations that accept S >= L admit
// multiple valid signatures for one message). Exact integer arithmetic only,
// no floats.
package eddsa

import (
	"errors"
	"math/big"
)

// groupOrderL is L = 2^252 + 27742317777372353535851937790883648493, the
// order of the edwards25519 base-point group (RFC 8032). Parsed once from
// its exact decimal value.
var groupOrderL = func() *big.Int {
	l, ok := new(big.Int).SetString("7237005577332262213973186563042994240857116359379907606001950938285454250989", 10)
	if !ok {
		panic("eddsa: could not parse group order L") // build-time-detectable authoring error
	}
	return l
}()

// ErrScalarOutOfRange is returned when S >= L (Step 5).
var ErrScalarOutOfRange = errors.New("eddsa: signature scalar S is out of range (S >= L)")

// leToBig decodes 32 little-endian octets as an unsigned big integer.
func leToBig(s [32]byte) *big.Int {
	be := make([]byte, 32)
	for i := 0; i < 32; i++ {
		be[i] = s[31-i]
	}
	return new(big.Int).SetBytes(be)
}

// checkScalarS runs EdDSA-Protodoc-1 Step 5: decode S as a little-endian
// 256-bit unsigned integer and reject if S >= L. It returns nil only when
// S < L. Uses exact big.Int comparison; no float arithmetic.
func checkScalarS(S [32]byte) error {
	if leToBig(S).Cmp(groupOrderL) >= 0 {
		return ErrScalarOutOfRange
	}
	return nil
}
