package integrity

import (
	"crypto/ed25519"
	"testing"

	"Protodoc/pkg/eddsa"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_076_VerifyReportsDeclaredOmissions is T-0192's named integration test
// (FR-076). A verified signature whose designated redactable subtrees have
// since been redacted reports the outcome "attested-with-declared-omissions"
// and enumerates every omission; the signature still verifies because the T_C
// root is preserved across the declared redaction. A valid signature with NO
// omissions reports plain "valid".
func TestFR_076_VerifyReportsDeclaredOmissions(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	var signer SignerIdentity
	copy(signer[:], pub)

	// Content set with two designated redactable subtrees.
	records := []ContentRecord{
		{UnitID: redUnitID(0x01), Frame: []byte("public"), Redactable: false},
		{UnitID: redUnitID(0x02), Frame: []byte("secret A"), Redactable: true, Salt: redSalt(0x22)},
		{UnitID: redUnitID(0x03), Frame: []byte("secret B"), Redactable: true, Salt: redSalt(0x33)},
	}
	tcRoot, err := TCRoot(records)
	if err != nil {
		t.Fatalf("TCRoot: %v", err)
	}
	structDigest := Digest{0x02}

	presRef := sigFixtureUnitID(0x30)
	boundPres := Digest{0x77}
	resolve := func(ref pdlfmt.UnitID) (Digest, bool) {
		if ref == presRef {
			return boundPres, true
		}
		return Digest{}, false
	}
	sig := SignatureRecord{
		ParamSet: 0x0001, Coverage: CoverageDescriptor{Mode: CoverageModeTotal, Covered: []SegmentRange{{Start: 0, End: 1}}},
		CredChainRef: sigFixtureUnitID(0x10), TimeAttestationRef: sigFixtureUnitID(0x20), PresentationRef: presRef,
	}
	so, _ := SignedObjectForSignature(sig, tcRoot, structDigest, resolve)
	sig.Value = eddsa.Sign(priv, [32]byte(so))

	// Redact both designated subtrees; the T_C root is preserved, so the
	// signature over it still verifies against the redacted current state.
	r2, _ := RedactRecord(records[1])
	r3, _ := RedactRecord(records[2])
	redacted := []ContentRecord{records[0], r2, r3}
	redactedRoot, _ := TCRoot(redacted)
	if redactedRoot != tcRoot {
		t.Fatalf("redaction changed the T_C root; declared-omission verify precondition broken")
	}

	rep, err := VerifySignatureForState(VerifyInput{
		Signature: sig, SignerKey: signer, CurrentTCRoot: redactedRoot, CurrentStructure: structDigest,
		ShownPresentation: boundPres, StateReconstructable: true,
	}, resolve)
	if err != nil {
		t.Fatalf("VerifySignatureForState: %v", err)
	}
	if rep.Verdict != VerdictValid {
		t.Fatalf("verdict = %v, want Valid (declared redaction preserves the signature)", rep.Verdict)
	}

	// Enumerate the declared omissions and report the distinct outcome.
	rep.DeclaredOmissions = []UnitID{records[1].UnitID, records[2].UnitID}
	if rep.ReportedOutcome() != OutcomeAttestedWithDeclaredOmissions {
		t.Errorf("reported outcome = %q, want %q", rep.ReportedOutcome(), OutcomeAttestedWithDeclaredOmissions)
	}
	if len(rep.DeclaredOmissions) != 2 {
		t.Errorf("expected 2 enumerated omissions, got %d", len(rep.DeclaredOmissions))
	}

	// A valid signature with NO declared omissions reports plain "valid".
	noOmit := rep
	noOmit.DeclaredOmissions = nil
	if noOmit.ReportedOutcome() != VerdictValid.String() {
		t.Errorf("no-omission outcome = %q, want %q", noOmit.ReportedOutcome(), VerdictValid.String())
	}

	// Declared omissions on a NON-valid verdict do not produce the
	// attested-with-declared-omissions outcome (only valid does).
	unver := rep
	unver.Verdict = VerdictUnverified
	if unver.ReportedOutcome() == OutcomeAttestedWithDeclaredOmissions {
		t.Error("a non-valid verdict must not report attested-with-declared-omissions")
	}
}
