package integrity

import "testing"

// TestN_COMPROMISE is T-0181's named conformance test (corpus N-COMPROMISE,
// FR-073). Negative corpus for the revocation-vs-signing-instant rule: every
// credential compromise recorded at or before the signing instant yields
// Unverified (zero false accepts), paired with strictly-after and
// no-compromise cases that yield Valid.
func TestN_COMPROMISE(t *testing.T) {
	const signing SigningInstant = 5_000_000

	// Negative corpus: compromise at or before signing -> Unverified.
	atOrBefore := []SigningInstant{0, 1, signing - 1, signing}
	for _, c := range atOrBefore {
		v := RevocationVerdict(signing, CompromiseInfo{Compromised: true, CompromiseTime: c}, true)
		if v != VerdictUnverified {
			t.Errorf("N-COMPROMISE: compromise at %d (signing %d) -> %v, want Unverified", c, signing, v)
		}
	}

	// Positive companions: compromise strictly after signing, or none, ->
	// Valid (a later compromise does not retroactively invalidate).
	after := []SigningInstant{signing + 1, signing + 1_000_000, 1 << 40}
	for _, c := range after {
		v := RevocationVerdict(signing, CompromiseInfo{Compromised: true, CompromiseTime: c}, true)
		if v != VerdictValid {
			t.Errorf("N-COMPROMISE: compromise at %d after signing -> %v, want Valid", c, v)
		}
	}
	if v := RevocationVerdict(signing, CompromiseInfo{Compromised: false}, true); v != VerdictValid {
		t.Errorf("N-COMPROMISE: no compromise -> %v, want Valid", v)
	}
}
