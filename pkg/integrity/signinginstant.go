// Signing-instant / time-attestation interval check (T-0178, FR-072). IF a
// signature's recorded signing instant lies OUTSIDE the interval attested by a
// time attestation whose own chain and revocation evidence verify offline,
// THEN a conforming reader presents the document as UNVERIFIED. The time
// attestation is only authoritative once its own evidence verifies offline
// (FR-071); an instant outside a verified interval is a genuine mismatch
// between what the signatory claims and what the timestamping authority
// vouched for.
package integrity

// SigningInstantVerdict returns the verification verdict contribution of the
// signing-instant-vs-attested-interval check (FR-072): VerdictValid when the
// signing instant lies within the attested interval, VerdictUnverified when it
// lies outside. The caller passes the interval the time attestation vouches
// for (extracted from the verified time attestation) and whether that time
// attestation's own evidence verified offline. If the time attestation did not
// verify offline, this check cannot make the signature Valid on its own and
// returns VerdictUnverified (an unverifiable time attestation cannot vouch for
// any instant).
func SigningInstantVerdict(signingInstant SigningInstant, attested TimeInterval, timeAttestationVerifiedOffline bool) Verdict {
	if !timeAttestationVerifiedOffline {
		return VerdictUnverified
	}
	if !attested.WellFormed() {
		return VerdictUnverified
	}
	if attested.Contains(signingInstant) {
		return VerdictValid
	}
	return VerdictUnverified
}
