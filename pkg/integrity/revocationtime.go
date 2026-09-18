// Revocation-vs-signing-instant check (T-0180, FR-073). IF revocation evidence
// records a credential compromise time AT OR BEFORE a signature's recorded
// signing instant, THEN a conforming reader presents the document as
// UNVERIFIED: a signature made with a credential already compromised at signing
// time cannot be trusted, even if it verifies cryptographically. The
// compromise time is extracted from the (offline-verified) revocation evidence
// by the caller; this check applies the at-or-before comparison.
package integrity

// CompromiseInfo describes what revocation evidence records about a credential.
type CompromiseInfo struct {
	// Compromised is true iff the revocation evidence records a compromise.
	Compromised bool
	// CompromiseTime is the recorded compromise time (meaningful only when
	// Compromised is true).
	CompromiseTime SigningInstant
}

// RevocationVerdict returns the verification verdict contribution of the
// revocation-vs-signing-instant check (FR-073): VerdictUnverified when the
// revocation evidence records a credential compromise at or before the signing
// instant, VerdictValid otherwise (no compromise recorded, or a compromise
// strictly after signing -- a later compromise does not retroactively
// invalidate a signature made while the credential was still good). The
// revocation evidence must have verified offline for its compromise record to
// be authoritative; a caller passes revocationVerifiedOffline=false to force
// Unverified when it could not establish that.
func RevocationVerdict(signingInstant SigningInstant, info CompromiseInfo, revocationVerifiedOffline bool) Verdict {
	if !revocationVerifiedOffline {
		return VerdictUnverified
	}
	if info.Compromised && info.CompromiseTime.AtOrBefore(signingInstant) {
		return VerdictUnverified
	}
	return VerdictValid
}
