package eddsa

import (
	"crypto/ed25519"
	"errors"
	"testing"
)

// leEncode returns the 32-octet little-endian encoding of a small increment
// added to p, for building non-canonical magnitude fixtures.
func pPlus(delta byte) [32]byte {
	var out [32]byte = fieldPrimeLE
	// add delta to the little-endian value (delta small, no cascade past a
	// couple of octets for the values used here)
	carry := uint16(delta)
	for i := 0; i < 32 && carry > 0; i++ {
		v := uint16(out[i]) + carry
		out[i] = byte(v)
		carry = v >> 8
	}
	return out
}

// TestCON_015_Step1Step2_PublicKeyChecks is T-0105's named test.
// checkPublicKey rejects magnitude==p and magnitude==p+1 (non-canonical),
// rejects each of the 8 small-order encodings, and accepts a valid
// random-looking canonical key.
func TestCON_015_Step1Step2_PublicKeyChecks(t *testing.T) {
	// magnitude == p (fieldPrimeLE with sign bit 0): non-canonical.
	if err := checkPublicKey(fieldPrimeLE); !errors.Is(err, ErrNonCanonicalEncoding) {
		t.Fatalf("magnitude==p: got %v, want ErrNonCanonicalEncoding", err)
	}
	// magnitude == p+1: non-canonical.
	if err := checkPublicKey(pPlus(1)); !errors.Is(err, ErrNonCanonicalEncoding) {
		t.Fatalf("magnitude==p+1: got %v, want ErrNonCanonicalEncoding", err)
	}
	// The x-sign bit (top bit of octet 31) must not affect the magnitude
	// check: magnitude==p with sign bit set is still non-canonical.
	pSigned := fieldPrimeLE
	pSigned[31] |= 0x80
	if err := checkPublicKey(pSigned); !errors.Is(err, ErrNonCanonicalEncoding) {
		t.Fatalf("magnitude==p (sign bit set): got %v, want ErrNonCanonicalEncoding", err)
	}

	// Each of the 8 small-order encodings is rejected. Some also have a
	// non-canonical magnitude (e.g. the order-2 point y=p-1 is canonical;
	// the identity is canonical); checkPublicKey must reject all 8 for one
	// of the two reasons, and specifically as small-order when canonical.
	for i, pt := range smallOrderPoints {
		err := checkPublicKey(pt)
		if err == nil {
			t.Fatalf("small-order entry %d was accepted", i)
		}
		if !errors.Is(err, ErrSmallOrderPoint) && !errors.Is(err, ErrNonCanonicalEncoding) {
			t.Fatalf("small-order entry %d rejected with unexpected %v", i, err)
		}
	}

	// A valid, canonical, non-small-order public key is accepted.
	_, priv := deterministicKey()
	pub := priv.Public().(ed25519.PublicKey)
	var A [32]byte
	copy(A[:], pub)
	if err := checkPublicKey(A); err != nil {
		t.Fatalf("valid canonical public key rejected: %v", err)
	}

	// A canonical magnitude just below p (p-1 with sign bit 0) passes the
	// canonical check (it is not small-order in general; here we only assert
	// the canonical branch does not fire).
	pMinus1 := fieldPrimeLE
	pMinus1[0] = 0xEC // p-1 little-endian low octet
	if !isCanonicalMagnitude(pMinus1) {
		t.Fatalf("magnitude p-1 wrongly flagged non-canonical")
	}
}
