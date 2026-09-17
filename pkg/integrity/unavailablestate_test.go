package integrity

import (
	"crypto/ed25519"
	"testing"

	"Protodoc/pkg/eddsa"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_062_UnavailableStateReportsDistinctVerdict is T-0159's named
// integration test (FR-062). IF a signature covers a state the current file
// can no longer reconstruct, the reader reports it as covering an UNAVAILABLE
// STATE -- a distinct verdict, never "invalid"/unverified and never "valid" --
// and never surfaces a signer identity for it.
func TestFR_062_UnavailableStateReportsDistinctVerdict(t *testing.T) {
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
		ParamSet:           0x0001,
		Coverage:           CoverageDescriptor{Mode: CoverageModeTotal, Covered: []SegmentRange{{Start: 0, End: 2}}},
		CredChainRef:       sigFixtureUnitID(0x10),
		TimeAttestationRef: sigFixtureUnitID(0x20),
		PresentationRef:    presRef,
	}
	so, _ := SignedObjectForSignature(sig, tcRoot, structDigest, resolve)
	sig.Value = eddsa.Sign(priv, [32]byte(so))
	sig.SignedObject = so

	// The signed state cannot be reconstructed (e.g. a lawful history trim /
	// NO_HISTORY post-migration): the verdict is UnavailableState, distinct
	// from both a cryptographic-failure Unverified and a Valid, EVEN THOUGH
	// the signature would otherwise verify against the (unavailable) state.
	in := VerifyInput{
		Signature: sig, SignerKey: signer,
		CurrentTCRoot: tcRoot, CurrentStructure: structDigest,
		ShownPresentation: boundPresentation, StateReconstructable: false,
	}
	rep, err := VerifySignatureForState(in, resolve)
	if err != nil {
		t.Fatalf("VerifySignatureForState: %v", err)
	}
	if rep.Verdict != VerdictUnavailableState {
		t.Fatalf("verdict = %v, want UnavailableState (FR-062)", rep.Verdict)
	}
	// Distinct from the other verdicts.
	if rep.Verdict == VerdictValid || rep.Verdict == VerdictUnverified {
		t.Errorf("unavailable state must not be reported as valid or unverified")
	}
	// No signer identity for a non-Valid verdict.
	if rep.Signer != nil {
		t.Errorf("unavailable-state verdict must not carry a signer identity")
	}
	if rep.Verdict.CarriesSignerIdentity() {
		t.Errorf("UnavailableState must not be an identity-carrying verdict")
	}
	// The report still names the pinned presentation and signed state.
	if rep.PinnedPresentation != boundPresentation {
		t.Errorf("report should still name the pinned presentation")
	}

	// Sanity: with the state reconstructable, the same signature IS Valid --
	// confirming UnavailableState is specifically the unreconstructable case,
	// not a general failure.
	in.StateReconstructable = true
	rep2, _ := VerifySignatureForState(in, resolve)
	if rep2.Verdict != VerdictValid {
		t.Fatalf("with reconstructable state the verdict = %v, want Valid", rep2.Verdict)
	}
}
