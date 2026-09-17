package integrity

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_065_MismatchedPresentationReportsUnattested is T-0156's named
// integration test (FR-065). A signature binds one presentation (its
// sig-presentation-ref). When a reader shows exactly that presentation, the
// presentation is attested; when it shows any other presentation, the reader
// reports it as NOT attested by that signature (the Unattested verdict) --
// distinct from valid and from a hard failure.
func TestFR_065_MismatchedPresentationReportsUnattested(t *testing.T) {
	presRef := sigFixtureUnitID(0x30)
	boundDigest := Digest{0xAA} // the presentation the signature bound

	sig := SignatureRecord{
		ParamSet:           1,
		Coverage:           CoverageDescriptor{Mode: CoverageModeTotal, Covered: []SegmentRange{{Start: 0, End: 1}}},
		CredChainRef:       sigFixtureUnitID(0x10),
		TimeAttestationRef: sigFixtureUnitID(0x20),
		PresentationRef:    presRef,
	}
	// The referenced segment currently has boundDigest as its slot digest.
	resolve := func(ref pdlfmt.UnitID) (Digest, bool) {
		if ref == presRef {
			return boundDigest, true
		}
		return Digest{}, false
	}

	// (1) Showing exactly the bound presentation -> attested.
	att, verdict, err := CheckPresentationAttestation(sig, boundDigest, resolve)
	if err != nil {
		t.Fatalf("attested case errored: %v", err)
	}
	if !att.Attested {
		t.Errorf("showing the bound presentation reported not-attested")
	}
	if verdict == VerdictUnattested {
		t.Errorf("showing the bound presentation reported Unattested")
	}

	// (2) Showing a DIFFERENT presentation (e.g. a reflowed view) -> the
	// verdict for that presentation is Unattested, not valid, not a failure.
	shown := Digest{0xBB} // a different presentation's slot digest
	att2, verdict2, err := CheckPresentationAttestation(sig, shown, resolve)
	if err != nil {
		t.Fatalf("mismatched case errored: %v", err)
	}
	if att2.Attested {
		t.Errorf("showing a different presentation reported attested")
	}
	if verdict2 != VerdictUnattested {
		t.Errorf("mismatched presentation verdict = %v, want Unattested", verdict2)
	}
	if att2.BoundDigest != boundDigest || att2.ShownDigest != shown {
		t.Errorf("attestation result did not carry the bound/shown digests")
	}
	// Unattested must not carry a signer identity (identity discipline: T-0164).
	if verdict2.CarriesSignerIdentity() {
		t.Error("Unattested verdict must not carry a signer identity")
	}

	// (3) An unresolved presentation ref is rejected (never silently attested).
	resolveNone := func(pdlfmt.UnitID) (Digest, bool) { return Digest{}, false }
	if _, v, err := CheckPresentationAttestation(sig, boundDigest, resolveNone); !errors.Is(err, ErrPresentationRefUnresolved) || v != VerdictUnattested {
		t.Errorf("unresolved ref: verdict=%v err=%v, want Unattested + ErrPresentationRefUnresolved", v, err)
	}
}
