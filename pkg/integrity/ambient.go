// Ambient-value allowlist (T-0165, NFR-005; CQ-004, a CP-004 constitutional
// invariant). Every value derived from wall-clock time, machine or user
// identity, filesystem metadata, process randomness, uninitialised padding, or
// editing-session history is EXCLUDED from the octet stream, EXCEPT at exactly
// four named sites: identifier minting, redaction commitment salts, signature
// values, and time attestations. An unenumerated exception is how ambient
// values leak back into content, so the set is closed and changeable only by
// constitution amendment, never by a format version.
package integrity

// AmbientSite is one of the four (and only four) sites where a value derived
// from a non-deterministic source is permitted to enter the octet stream.
type AmbientSite int

const (
	// AmbientIdentifierMinting: run_id / unit-id minting via crypto/rand
	// (pkg/content/mint).
	AmbientIdentifierMinting AmbientSite = iota
	// AmbientRedactionSalt: redaction commitment salts (RedactionCommitment).
	AmbientRedactionSalt
	// AmbientSignatureValue: signature values (sig-value; the key material is
	// the only non-deterministic input, RFC 8032 signing being deterministic).
	AmbientSignatureValue
	// AmbientTimeAttestation: time attestations (ATTESTATION_EVIDENCE,
	// ae-kind = time-attestation).
	AmbientTimeAttestation
)

func (s AmbientSite) String() string {
	switch s {
	case AmbientIdentifierMinting:
		return "identifier-minting"
	case AmbientRedactionSalt:
		return "redaction-commitment-salt"
	case AmbientSignatureValue:
		return "signature-value"
	case AmbientTimeAttestation:
		return "time-attestation"
	default:
		return "unknown"
	}
}

// AllowlistedAmbientSites is the CLOSED set of the four (and only four) sites
// where an ambient-derived value may enter the octet stream (NFR-005).
var AllowlistedAmbientSites = []AmbientSite{
	AmbientIdentifierMinting, AmbientRedactionSalt, AmbientSignatureValue, AmbientTimeAttestation,
}

// IsAllowlistedAmbientSite reports whether s is one of the four named sites.
func IsAllowlistedAmbientSite(s AmbientSite) bool {
	return s >= AmbientIdentifierMinting && s <= AmbientTimeAttestation
}
