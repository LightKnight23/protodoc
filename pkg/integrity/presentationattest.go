// Presentation attestation (T-0156, FR-065; integrity.abnf S3.2, document.abnf
// S7.4). A signature binds exactly one presentation: the PRESENTATION_ARTEFACT
// its sig-presentation-ref names, whose slot digest is committed into
// signed_object. IF a reader presents a signed document in any presentation
// OTHER than that bound one, it must report the presentation as NOT ATTESTED
// by that signature -- the bytes shown were not the bytes signed. This is a
// distinct status, neither a hard failure nor "valid": without it the
// reflowable view would inherit the signer's name and "what did the signatory
// see" would have no answer.
package integrity

// PresentationAttestation is the result of checking whether a shown
// presentation is attested by a signature.
type PresentationAttestation struct {
	// Attested is true iff the shown presentation is the one the signature
	// bound (their slot digests match).
	Attested bool
	// BoundDigest is the presentation slot digest the signature committed to.
	BoundDigest Digest
	// ShownDigest is the slot digest of the presentation actually shown.
	ShownDigest Digest
}

// CheckPresentationAttestation compares the presentation actually shown
// (shownPresentationDigest, the slot digest of the PRESENTATION_ARTEFACT a
// reader rendered) against the one the signature bound (the slot digest of the
// segment named by sig-presentation-ref, resolved fresh). They match iff the
// reader is showing exactly the signed presentation. It returns the attestation
// result and, when the shown presentation is not the bound one, the Unattested
// verdict for that presentation (FR-065).
func CheckPresentationAttestation(sig SignatureRecord, shownPresentationDigest Digest, resolve SlotDigestResolver) (PresentationAttestation, Verdict, error) {
	bound, ok := resolve(sig.PresentationRef)
	if !ok {
		return PresentationAttestation{}, VerdictUnattested, ErrPresentationRefUnresolved
	}
	att := PresentationAttestation{
		Attested:    bound == shownPresentationDigest,
		BoundDigest: bound,
		ShownDigest: shownPresentationDigest,
	}
	if att.Attested {
		// The shown presentation IS the bound one; this check does not by
		// itself make the signature valid (the cryptographic verify does), so
		// it returns the neutral Valid-eligible signal by leaving the verdict
		// to the caller. We return VerdictValid to mean "attested here", but
		// callers combine it with the cryptographic verdict.
		return att, VerdictValid, nil
	}
	// The bytes shown were not the bytes signed: not attested by this signature.
	return att, VerdictUnattested, nil
}
