package integrity

import "testing"

// TestN_TIME is T-0179's named conformance test (corpus N-TIME, FR-072). It is
// the negative corpus for the signing-instant-vs-attested-interval rule: a set
// of signing instants that lie outside a verified time attestation's interval,
// each of which must yield Unverified, paired with in-interval instants that
// yield Valid -- zero false accepts.
func TestN_TIME(t *testing.T) {
	iv := TimeInterval{NotBefore: 1_000_000, NotAfter: 2_000_000}

	// Negative corpus: instants outside the interval, all Unverified.
	outside := []SigningInstant{
		0, 1, 999_999, // before NotBefore
		2_000_001, 3_000_000, 1 << 40, // after NotAfter
	}
	for _, s := range outside {
		if v := SigningInstantVerdict(s, iv, true); v != VerdictUnverified {
			t.Errorf("N-TIME: signing instant %d outside [%d,%d] -> %v, want Unverified", s, iv.NotBefore, iv.NotAfter, v)
		}
	}

	// Positive companions: instants inside (and at the closed boundaries),
	// all Valid -- so the corpus is not vacuously all-reject.
	inside := []SigningInstant{iv.NotBefore, 1_500_000, iv.NotAfter}
	for _, s := range inside {
		if v := SigningInstantVerdict(s, iv, true); v != VerdictValid {
			t.Errorf("N-TIME: signing instant %d inside interval -> %v, want Valid", s, v)
		}
	}

	// Every outside instant is Unverified even at a nanosecond boundary just
	// past the edge (off-by-one closes the boundary correctly).
	if v := SigningInstantVerdict(iv.NotAfter+1, iv, true); v != VerdictUnverified {
		t.Errorf("N-TIME: NotAfter+1 -> %v, want Unverified", v)
	}
	if v := SigningInstantVerdict(iv.NotBefore-1, iv, true); v != VerdictUnverified {
		t.Errorf("N-TIME: NotBefore-1 -> %v, want Unverified", v)
	}
}
