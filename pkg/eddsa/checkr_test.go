package eddsa

import (
	"errors"
	"testing"
)

// TestCON_015_Step3Step4_RChecks is T-0106's named test. checkR rejects the
// same non-canonical and small-order encoding classes as checkPublicKey,
// exercised table-driven, and accepts a canonical non-small-order R.
func TestCON_015_Step3Step4_RChecks(t *testing.T) {
	cases := []struct {
		name    string
		R       [32]byte
		wantErr error // nil = accept
	}{
		{"magnitude==p (non-canonical)", fieldPrimeLE, ErrNonCanonicalEncoding},
		{"magnitude==p+1 (non-canonical)", pPlus(1), ErrNonCanonicalEncoding},
	}
	for _, c := range cases {
		err := checkR(c.R)
		if c.wantErr == nil {
			if err != nil {
				t.Fatalf("%s: got %v, want accept", c.name, err)
			}
			continue
		}
		if !errors.Is(err, c.wantErr) {
			t.Fatalf("%s: got %v, want %v", c.name, err, c.wantErr)
		}
	}

	// All 8 small-order encodings are rejected (small-order or, if the
	// magnitude is non-canonical, that).
	for i, pt := range smallOrderPoints {
		err := checkR(pt)
		if err == nil {
			t.Fatalf("small-order entry %d accepted as R", i)
		}
		if !errors.Is(err, ErrSmallOrderPoint) && !errors.Is(err, ErrNonCanonicalEncoding) {
			t.Fatalf("small-order entry %d as R rejected with unexpected %v", i, err)
		}
	}

	// checkR and checkPublicKey are the same check: they agree on every
	// small-order table entry and on the p/p+1 fixtures.
	for _, R := range append([][32]byte{fieldPrimeLE, pPlus(1)}, smallOrderPoints[:]...) {
		if (checkR(R) == nil) != (checkPublicKey(R) == nil) {
			t.Fatalf("checkR and checkPublicKey disagree on an encoding")
		}
	}

	// A canonical, non-small-order R (a real signature's R half) is
	// accepted. Produce one by signing with the deterministic key.
	_, priv := deterministicKey()
	var msg [32]byte
	msg[0] = 0x11
	sig := Sign(priv, msg)
	var R [32]byte
	copy(R[:], sig[:32])
	if err := checkR(R); err != nil {
		t.Fatalf("valid signature R rejected: %v", err)
	}
}
