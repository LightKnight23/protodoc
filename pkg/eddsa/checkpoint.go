// EdDSA-Protodoc-1 Steps 1-2 and 3-4: the canonical-encoding and
// small-order-point checks applied to the public key A and the signature
// component R (integrity.abnf S6). These are curve-arithmetic-free byte and
// integer comparisons that pre-filter the exact inputs on which EdDSA
// implementations are documented to diverge; no curve point is decoded here.
package eddsa

import (
	"errors"
)

// fieldPrimeLE is p = 2^255 - 19 as 32 little-endian octets. A canonical
// Ed25519 point encoding's 255-bit y-magnitude (the encoding with the top
// bit of octet 31 masked off) must be strictly less than p; a magnitude >= p
// is a NON-CANONICAL encoding (RFC 8032 section 5.1.3), which
// EdDSA-Protodoc-1 rejects.
var fieldPrimeLE = [32]byte{
	0xED, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F,
}

var (
	// ErrNonCanonicalEncoding is returned when a point encoding's 255-bit
	// magnitude is >= p (non-canonical; Step 1 / Step 3).
	ErrNonCanonicalEncoding = errors.New("eddsa: non-canonical point encoding (255-bit magnitude >= p)")
	// ErrSmallOrderPoint is returned when a point encoding matches one of
	// the 8 small-order table entries (Step 2 / Step 4).
	ErrSmallOrderPoint = errors.New("eddsa: small-order point encoding rejected")
)

// isCanonicalMagnitude reports whether enc's 255-bit magnitude (top bit of
// octet 31 masked off) is strictly less than p. A and R are public values,
// so a plain (not constant-time) little-endian magnitude compare is used:
// the timing side channel EdDSA-Protodoc-1 closes is between "rejected at an
// early step" and "rejected at Step 7", handled by Verify's short-circuit,
// not by this comparison of public bytes.
func isCanonicalMagnitude(enc [32]byte) bool {
	var mag [32]byte = enc
	mag[31] &= 0x7F // mask off the x-sign bit
	// Compare mag < p, little-endian: scan from the most-significant octet.
	for i := 31; i >= 0; i-- {
		if mag[i] < fieldPrimeLE[i] {
			return true // strictly less at the first differing octet
		}
		if mag[i] > fieldPrimeLE[i] {
			return false
		}
	}
	return false // exactly equal to p => non-canonical
}

// isSmallOrderEncoding reports whether enc exactly matches any of the 8
// small-order table entries.
func isSmallOrderEncoding(enc [32]byte) bool {
	for _, pt := range smallOrderPoints {
		if enc == pt {
			return true
		}
	}
	return false
}

// checkPublicKey runs EdDSA-Protodoc-1 Steps 1-2 on the public key A: reject
// a non-canonical 255-bit magnitude (>= p), then reject a small-order
// encoding. It returns nil only when A passes both.
func checkPublicKey(A [32]byte) error {
	if !isCanonicalMagnitude(A) {
		return ErrNonCanonicalEncoding
	}
	if isSmallOrderEncoding(A) {
		return ErrSmallOrderPoint
	}
	return nil
}
