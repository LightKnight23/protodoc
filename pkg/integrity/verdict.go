// Verification verdict (T-0156, FR-065; the closed four-value verdict enum is asserted by T-0157, data-model.md S7 step
// 10, cli.md S5). Verification status is reported as a value DISTINCT from
// extracted content, taking one of a CLOSED set of exactly four
// verdicts:
//
//   - Valid: the signature verifies against the current state's fresh
//     signed_object, and the presentation shown is the one it bound.
//   - Unverified: a requested cryptographic check produced a negative result
//     (EdDSA-Protodoc-1 failure, unsupported time attestation, revocation
//     before signing, an undeclared omission).
//   - UnavailableState: the signature covers a state the current file can no
//     longer reconstruct (FR-062) -- never "failed", never "valid".
//   - Unattested: the presentation actually shown is not the one bound by the
//     signature (FR-065) -- the bytes a reader sees were not the bytes signed;
//     a distinct status, not a hard failure and not valid.
//
// A signer identity is only ever displayed alongside Valid (this identity
// discipline is asserted by T-0164,
// cli.md non-negotiable #4).
package integrity

// Verdict is the closed four-value verification verdict (asserted by T-0157).
type Verdict int

const (
	VerdictValid Verdict = iota
	VerdictUnverified
	VerdictUnavailableState
	VerdictUnattested
)

// verdictNames is the canonical spelling of each verdict for reports. Its
// length is exactly the size of the closed set, so a test can assert the
// enum has not silently grown.
var verdictNames = map[Verdict]string{
	VerdictValid:            "valid",
	VerdictUnverified:       "unverified",
	VerdictUnavailableState: "unavailable_state",
	VerdictUnattested:       "unattested",
}

// AllVerdicts is the closed set of every defined verdict, in enum order.
var AllVerdicts = []Verdict{VerdictValid, VerdictUnverified, VerdictUnavailableState, VerdictUnattested}

func (v Verdict) String() string {
	if s, ok := verdictNames[v]; ok {
		return s
	}
	return "undefined"
}

// IsDefined reports whether v is one of the closed four verdicts.
func (v Verdict) IsDefined() bool {
	_, ok := verdictNames[v]
	return ok
}

// CarriesSignerIdentity reports whether a signer identity may be displayed
// alongside this verdict. Only Valid may (asserted by T-0164; cli.md non-negotiable #4).
func (v Verdict) CarriesSignerIdentity() bool {
	return v == VerdictValid
}
