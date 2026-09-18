package integrity

import "testing"

// TestFR_072_SigningInstantOutsideIntervalUnverified is T-0178's named
// integration test (FR-072). When a signature's recorded signing instant lies
// outside the interval attested by a time attestation whose own chain and
// revocation verify offline, the document is presented as UNVERIFIED; an
// instant within the verified interval is Valid; and an instant against a time
// attestation that did NOT verify offline is Unverified regardless of the
// instant.
func TestFR_072_SigningInstantOutsideIntervalUnverified(t *testing.T) {
	interval := TimeInterval{NotBefore: 1000, NotAfter: 2000}

	// Inside the verified interval -> Valid.
	if v := SigningInstantVerdict(1500, interval, true); v != VerdictValid {
		t.Errorf("instant inside verified interval: verdict = %v, want Valid", v)
	}
	// Exactly at the boundaries (closed interval) -> Valid.
	if v := SigningInstantVerdict(1000, interval, true); v != VerdictValid {
		t.Errorf("instant at NotBefore: verdict = %v, want Valid", v)
	}
	if v := SigningInstantVerdict(2000, interval, true); v != VerdictValid {
		t.Errorf("instant at NotAfter: verdict = %v, want Valid", v)
	}

	// Outside the verified interval (before and after) -> Unverified.
	if v := SigningInstantVerdict(999, interval, true); v != VerdictUnverified {
		t.Errorf("instant before interval: verdict = %v, want Unverified", v)
	}
	if v := SigningInstantVerdict(2001, interval, true); v != VerdictUnverified {
		t.Errorf("instant after interval: verdict = %v, want Unverified", v)
	}

	// Time attestation did not verify offline -> Unverified even for an
	// instant that would otherwise be within the interval.
	if v := SigningInstantVerdict(1500, interval, false); v != VerdictUnverified {
		t.Errorf("instant within an UNVERIFIED time attestation: verdict = %v, want Unverified", v)
	}

	// A malformed (inverted) interval -> Unverified.
	inverted := TimeInterval{NotBefore: 2000, NotAfter: 1000}
	if v := SigningInstantVerdict(1500, inverted, true); v != VerdictUnverified {
		t.Errorf("instant against an inverted interval: verdict = %v, want Unverified", v)
	}
}
