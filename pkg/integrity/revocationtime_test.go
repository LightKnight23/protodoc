package integrity

import "testing"

// TestFR_073_RevocationAtOrBeforeSigningUnverified is T-0180's named
// integration test (FR-073). When revocation evidence (verified offline)
// records a credential compromise AT OR BEFORE the signing instant, the
// document is Unverified; a compromise strictly AFTER signing does not
// retroactively invalidate the signature; and revocation evidence that did not
// verify offline forces Unverified.
func TestFR_073_RevocationAtOrBeforeSigningUnverified(t *testing.T) {
	const signing SigningInstant = 1000

	// Compromise strictly before signing -> Unverified.
	if v := RevocationVerdict(signing, CompromiseInfo{Compromised: true, CompromiseTime: 900}, true); v != VerdictUnverified {
		t.Errorf("compromise before signing: verdict = %v, want Unverified", v)
	}
	// Compromise exactly AT signing -> Unverified (at-or-before).
	if v := RevocationVerdict(signing, CompromiseInfo{Compromised: true, CompromiseTime: 1000}, true); v != VerdictUnverified {
		t.Errorf("compromise at signing instant: verdict = %v, want Unverified", v)
	}
	// Compromise strictly after signing -> Valid (does not retroactively
	// invalidate a signature made while the credential was still good).
	if v := RevocationVerdict(signing, CompromiseInfo{Compromised: true, CompromiseTime: 1001}, true); v != VerdictValid {
		t.Errorf("compromise after signing: verdict = %v, want Valid", v)
	}
	// No compromise recorded -> Valid.
	if v := RevocationVerdict(signing, CompromiseInfo{Compromised: false}, true); v != VerdictValid {
		t.Errorf("no compromise: verdict = %v, want Valid", v)
	}
	// Revocation evidence that did not verify offline -> Unverified, even with
	// no compromise recorded (an unverifiable revocation cannot vouch).
	if v := RevocationVerdict(signing, CompromiseInfo{Compromised: false}, false); v != VerdictUnverified {
		t.Errorf("unverified revocation evidence: verdict = %v, want Unverified", v)
	}
}
