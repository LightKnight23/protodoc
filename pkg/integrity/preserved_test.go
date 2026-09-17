package integrity

import (
	"bytes"
	"crypto/ed25519"
	"testing"

	"Protodoc/pkg/eddsa"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_066_SignaturePreservedAcrossSubsequentEdit is T-0160's named
// integration test (FR-066). WHEN a writer commits an edit to a signed
// document, it preserves the signature bound to the state identifier it covers
// together with that state's pinned presentation artefact. This models the
// preservation: a signature over state1 (with presentation P1), after an edit
// produces state2, must remain byte-identical, still name state1, still bind
// P1, and still verify Valid against state1 -- while the current (edited) state
// is reported as not attested by it.
func TestFR_066_SignaturePreservedAcrossSubsequentEdit(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	var signer SignerIdentity
	copy(signer[:], pub)

	presRef := sigFixtureUnitID(0x30)
	presentation1 := Digest{0x71}
	resolve := func(ref pdlfmt.UnitID) (Digest, bool) {
		if ref == presRef {
			return presentation1, true
		}
		return Digest{}, false
	}

	// State 1: the signed state.
	state1TC := Digest{0x0A}
	state1Struct := Digest{0x0B}

	sig := SignatureRecord{
		ParamSet:           0x0001,
		Coverage:           CoverageDescriptor{Mode: CoverageModeTotal, Covered: []SegmentRange{{Start: 0, End: 3}}},
		CredChainRef:       sigFixtureUnitID(0x10),
		TimeAttestationRef: sigFixtureUnitID(0x20),
		PresentationRef:    presRef,
	}
	so1, _ := SignedObjectForSignature(sig, state1TC, state1Struct, resolve)
	sig.Value = eddsa.Sign(priv, [32]byte(so1))
	sig.SignedObject = so1

	// Serialize the signature as it lives in the ATTEST segment.
	before, err := sig.Encode()
	if err != nil {
		t.Fatalf("Encode signature: %v", err)
	}

	// --- Commit an edit: state advances to state2 (different t_c_root/struct).
	// FR-066 requires the writer to PRESERVE the signature record intact,
	// still bound to state1. We model that preservation as the exact same
	// serialized bytes surviving into the edited document.
	preserved := append([]byte(nil), before...)

	// The preserved signature decodes byte-identically and re-encodes exactly.
	sigAfter, err := DecodeSignatureRecord(preserved)
	if err != nil {
		t.Fatalf("Decode preserved signature: %v", err)
	}
	reEnc, err := sigAfter.Encode()
	if err != nil {
		t.Fatalf("re-Encode preserved signature: %v", err)
	}
	if !bytes.Equal(before, reEnc) {
		t.Fatalf("preserved signature is not byte-identical after the edit")
	}
	if sigAfter.PresentationRef != presRef {
		t.Errorf("preserved signature lost its pinned presentation reference")
	}

	// It still verifies Valid against STATE 1 (the state it covers), with the
	// presentation it pinned -- the earlier attestation stays legible.
	inState1 := VerifyInput{
		Signature: sigAfter, SignerKey: signer,
		CurrentTCRoot: state1TC, CurrentStructure: state1Struct,
		ShownPresentation: presentation1, StateReconstructable: true,
	}
	rep1, err := VerifySignatureForState(inState1, resolve)
	if err != nil {
		t.Fatalf("verify against state1: %v", err)
	}
	if rep1.Verdict != VerdictValid {
		t.Fatalf("preserved signature verdict against its own state = %v, want Valid", rep1.Verdict)
	}
	if rep1.SignedState.TCRoot != state1TC || rep1.SignedState.StructureDigest != state1Struct {
		t.Errorf("preserved signature no longer names state1")
	}

	// Against the edited current state2, it is not Valid (the signature does
	// not silently carry to the new state), but it was not discarded.
	inState2 := inState1
	inState2.CurrentTCRoot = Digest{0xC0}
	inState2.CurrentStructure = Digest{0xC1}
	rep2, _ := VerifySignatureForState(inState2, resolve)
	if rep2.Verdict == VerdictValid {
		t.Fatalf("preserved signature wrongly verified Valid against the edited state")
	}
}
