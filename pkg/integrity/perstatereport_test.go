package integrity

import (
	"crypto/ed25519"
	"testing"

	"Protodoc/pkg/eddsa"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_067_PerStateAttestationReportCoversAllOutcomes is T-0161's named
// integration test (FR-067). A per-state attestation report must be able to
// express every one of the four verdict outcomes -- valid,
// unavailable_state, unattested, unverified -- each naming the signed state
// and pinned presentation, and surfacing the signer identity ONLY for valid.
// This drives the same signature through all four situations and asserts the
// full outcome coverage.
func TestFR_067_PerStateAttestationReportCoversAllOutcomes(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	var signer SignerIdentity
	copy(signer[:], pub)

	presRef := sigFixtureUnitID(0x30)
	boundPresentation := Digest{0x71}
	resolve := func(ref pdlfmt.UnitID) (Digest, bool) {
		if ref == presRef {
			return boundPresentation, true
		}
		return Digest{}, false
	}
	tcRoot := Digest{0x0A}
	structDigest := Digest{0x0B}

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

	base := VerifyInput{
		Signature: sig, SignerKey: signer,
		CurrentTCRoot: tcRoot, CurrentStructure: structDigest,
		ShownPresentation: boundPresentation, StateReconstructable: true,
	}

	// Four situations -> the four verdicts.
	valid := base

	unavailable := base
	unavailable.StateReconstructable = false

	unattested := base
	unattested.ShownPresentation = Digest{0xEE} // a different presentation

	unverified := base
	unverified.CurrentStructure = Digest{0x99} // edited state -> signed_object differs -> crypto fails

	cases := []struct {
		name string
		in   VerifyInput
		want Verdict
	}{
		{"valid", valid, VerdictValid},
		{"unavailable", unavailable, VerdictUnavailableState},
		{"unattested", unattested, VerdictUnattested},
		{"unverified", unverified, VerdictUnverified},
	}

	seen := map[Verdict]bool{}
	for _, c := range cases {
		rep, err := VerifySignatureForState(c.in, resolve)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if rep.Verdict != c.want {
			t.Errorf("%s: verdict = %v, want %v", c.name, rep.Verdict, c.want)
		}
		seen[rep.Verdict] = true
		// Every report names the pinned presentation.
		if rep.PinnedPresentation != boundPresentation {
			t.Errorf("%s: report does not name the pinned presentation", c.name)
		}
		// Signer identity ONLY for valid.
		if c.want == VerdictValid {
			if rep.Signer == nil {
				t.Errorf("%s: valid report must name the signer", c.name)
			}
		} else if rep.Signer != nil {
			t.Errorf("%s: non-valid report must not name a signer", c.name)
		}
	}

	// All four outcomes are covered by the per-state report machinery.
	for _, v := range AllVerdicts {
		if !seen[v] {
			t.Errorf("per-state report did not cover verdict outcome %v", v)
		}
	}
}
