// Per-state attestation reporting (T-0158, FR-067; data-model.md S7 step 10,
// cli.md S5). A reader reports attestation PER STATE: it names the signed
// state, its signer and its pinned presentation, and reports the CURRENT state
// as not attested when it differs from a signed state. Applied naively, the
// first keystroke after signing turns a correctly signed contract into an
// unsigned one with no way to say what was signed; per-state reporting keeps
// the earlier attestation legible without ever showing a signer's name over
// edited content.
package integrity

import "Protodoc/pkg/eddsa"

// ActedUponWithinCoverage reports whether every ordinal in the acted-upon set
// lies within one of the signature's covered ranges (FR-115). If ANY
// acted-upon octet lies outside every covered range, a reader must present the
// document as unverified and show no signer identity -- a positive indicator
// over partial coverage is worse than none. This does not consult the
// uncovered list: coverage for this purpose is strictly the covered ranges.
func ActedUponWithinCoverage(cov CoverageDescriptor, actedUpon []uint64) bool {
	within := func(o uint64) bool {
		for _, r := range cov.Covered {
			if o >= r.Start && o < r.End {
				return true
			}
		}
		return false
	}
	for _, o := range actedUpon {
		if !within(o) {
			return false
		}
	}
	return true
}

// SignerIdentity is the signer's public verifying key. It is only ever
// surfaced in a report alongside a Valid verdict (identity discipline
// conformance-tested by T-0164).
type SignerIdentity [32]byte

// SignedState names the state a signature attests: the (t_c_root,
// structure_digest) pair at signing time is the signature identity
// (integrity.abnf S5), which together pin the exact signed content and
// structure.
type SignedState struct {
	TCRoot          Digest
	StructureDigest Digest
}

// AttestationReport is verify's per-signature, per-state report (FR-067). It
// names the signed state, the pinned presentation, and the verdict; the signer
// identity is populated ONLY when the verdict is Valid (see T-0164).
type AttestationReport struct {
	SignedState         SignedState
	PinnedPresentation  Digest          // the bound presentation slot digest
	Verdict             Verdict         // valid / unverified / unavailable_state / unattested
	Signer              *SignerIdentity // non-nil ONLY when Verdict == Valid
	CurrentStateMatches bool            // whether the current state equals the signed state
	// DeclaredOmissions lists every redactable subtree the signature
	// designated that has since been redacted (a DECLARED omission). When the
	// verdict is Valid and this list is non-empty, the reported outcome is
	// "attested-with-declared-omissions" (FR-076): the signature still
	// verifies (the T_C root is preserved across a declared redaction) and
	// every omission is enumerated. An UNDECLARED omission (FR-077) yields
	// Unverified and never populates this list.
	DeclaredOmissions []UnitID
}

// OutcomeAttestedWithDeclaredOmissions is the reported outcome string when a
// valid signature has designated omissions (FR-076, spec.md).
const OutcomeAttestedWithDeclaredOmissions = "attested-with-declared-omissions"

// ReportedOutcome returns the human outcome for a report: a Valid verdict with
// one or more declared omissions reports "attested-with-declared-omissions"
// (FR-076); otherwise the verdict's own name. This distinct outcome is a
// Valid-family outcome (the signature verifies) carried as Valid + a non-empty
// omission list rather than a fifth verdict enum member -- the closed verdict
// set stays four, with cli.md S5's per-signature verdict field
// (valid/unverified/unavailable_state) plus a separate declared_omissions
// array.
func (r AttestationReport) ReportedOutcome() string {
	if r.Verdict == VerdictValid && len(r.DeclaredOmissions) > 0 {
		return OutcomeAttestedWithDeclaredOmissions
	}
	return r.Verdict.String()
}

// VerifyInput carries what a reader observes about the CURRENT file to verify
// a signature against the state it attests.
type VerifyInput struct {
	Signature            SignatureRecord
	SignerKey            SignerIdentity // the public key the signature is checked against
	CurrentTCRoot        Digest         // fresh T_C_root over the covered subtree at verify time
	CurrentStructure     Digest         // fresh structure_digest
	ShownPresentation    Digest         // the presentation slot digest the reader is showing
	StateReconstructable bool           // false if the signed state cannot be reconstructed (FR-062)
}

// VerifySignatureForState produces the per-state AttestationReport for one
// signature (FR-067). Ordering of verdict determination:
//
//  1. If the signed state cannot be reconstructed, the verdict is
//     UnavailableState (FR-062) -- never valid, never a failure.
//  2. Else if the shown presentation is not the bound one, the verdict is
//     Unattested (FR-065) -- the current presentation is not attested.
//  3. Else recompute signed_object from the CURRENT state and run
//     EdDSA-Protodoc-1; a negative result is Unverified, a positive result is
//     Valid.
//
// The signer identity is attached ONLY to a Valid verdict (T-0164). The report
// always names the signed state and the pinned presentation regardless of
// verdict, so an earlier attestation stays legible.
func VerifySignatureForState(in VerifyInput, resolve SlotDigestResolver) (AttestationReport, error) {
	bound, ok := resolve(in.Signature.PresentationRef)
	if !ok {
		return AttestationReport{}, ErrPresentationRefUnresolved
	}
	rep := AttestationReport{
		SignedState:        SignedState{TCRoot: in.CurrentTCRoot, StructureDigest: in.CurrentStructure},
		PinnedPresentation: bound,
	}

	// (1) Unavailable state (FR-062).
	if !in.StateReconstructable {
		rep.Verdict = VerdictUnavailableState
		return rep, nil
	}

	// (2) Presentation attestation (FR-065): the shown presentation must be
	// the bound one.
	if bound != in.ShownPresentation {
		rep.Verdict = VerdictUnattested
		return rep, nil
	}

	// (3) Cryptographic verify over the freshly recomputed signed_object.
	so, err := SignedObjectForSignature(in.Signature, in.CurrentTCRoot, in.CurrentStructure, resolve)
	if err != nil {
		return AttestationReport{}, err
	}
	ok = eddsa.VerifyWithParamSet(eddsa.ParamSet(in.Signature.ParamSet), [32]byte(in.SignerKey), [32]byte(so), in.Signature.Value)
	if !ok {
		rep.Verdict = VerdictUnverified
		return rep, nil
	}
	rep.Verdict = VerdictValid
	rep.CurrentStateMatches = true
	// Signer identity is surfaced ONLY for a Valid verdict (T-0164).
	signer := in.SignerKey
	rep.Signer = &signer
	return rep, nil
}
