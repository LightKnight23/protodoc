package integrity

import (
	"crypto/ed25519"
	"testing"

	"Protodoc/pkg/eddsa"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_067_VerifyReportsSignedStateSignerAndPresentation is T-0158's named
// integration test (FR-063/FR-067). It signs a real signed_object with an
// Ed25519 key and verifies through the per-state orchestration, asserting the
// AttestationReport names the signed state (t_c_root, structure_digest), the
// pinned presentation, and -- only for a Valid verdict -- the signer identity.
// It then confirms an edited current state no longer verifies as Valid.
func TestFR_067_VerifyReportsSignedStateSignerAndPresentation(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	var signer SignerIdentity
	copy(signer[:], pub)

	presRef := sigFixtureUnitID(0x30)
	boundPresentation := Digest{0x77}
	resolve := func(ref pdlfmt.UnitID) (Digest, bool) {
		if ref == presRef {
			return boundPresentation, true
		}
		return Digest{}, false
	}

	tcRoot := Digest{0x01}
	structDigest := Digest{0x02}

	sig := SignatureRecord{
		ParamSet:           0x0001, // Ed25519 per the S10.2 allowlist
		Coverage:           CoverageDescriptor{Mode: CoverageModeTotal, Covered: []SegmentRange{{Start: 0, End: 2}}},
		CredChainRef:       sigFixtureUnitID(0x10),
		TimeAttestationRef: sigFixtureUnitID(0x20),
		PresentationRef:    presRef,
	}

	// Sign the signed_object computed over the signed state.
	so, err := SignedObjectForSignature(sig, tcRoot, structDigest, resolve)
	if err != nil {
		t.Fatalf("SignedObjectForSignature: %v", err)
	}
	sig.Value = eddsa.Sign(priv, [32]byte(so))
	sig.SignedObject = so

	// (1) Verify against the signed state, showing the bound presentation.
	in := VerifyInput{
		Signature: sig, SignerKey: signer,
		CurrentTCRoot: tcRoot, CurrentStructure: structDigest,
		ShownPresentation: boundPresentation, StateReconstructable: true,
	}
	rep, err := VerifySignatureForState(in, resolve)
	if err != nil {
		t.Fatalf("VerifySignatureForState: %v", err)
	}
	if rep.Verdict != VerdictValid {
		t.Fatalf("verdict = %v, want Valid", rep.Verdict)
	}
	if rep.SignedState.TCRoot != tcRoot || rep.SignedState.StructureDigest != structDigest {
		t.Errorf("report does not name the signed state")
	}
	if rep.PinnedPresentation != boundPresentation {
		t.Errorf("report does not name the pinned presentation")
	}
	if rep.Signer == nil || *rep.Signer != signer {
		t.Errorf("Valid report must name the signer identity")
	}
	if !rep.CurrentStateMatches {
		t.Errorf("Valid report should mark the current state as matching")
	}

	// (2) An edited current state (different structure_digest) recomputes a
	// different signed_object, so the same signature no longer verifies as
	// Valid; the verdict becomes Unverified and NO signer identity is shown.
	edited := in
	edited.CurrentStructure = Digest{0x99}
	rep2, err := VerifySignatureForState(edited, resolve)
	if err != nil {
		t.Fatalf("VerifySignatureForState (edited): %v", err)
	}
	if rep2.Verdict == VerdictValid {
		t.Fatalf("edited state verified as Valid; the signature must not carry to a changed state")
	}
	if rep2.Signer != nil {
		t.Errorf("a non-Valid verdict must not name a signer identity")
	}
	// The report still names the (new current) state and the pinned
	// presentation, keeping the attestation context legible.
	if rep2.PinnedPresentation != boundPresentation {
		t.Errorf("report should still name the pinned presentation")
	}
}
