package integrity

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_064_SignedObjectSourcesPresentationDigestFromReferencedSlot is
// T-0155's named unit test (FR-064). The presentation_artefact_digest input to
// signed_object MUST be the referenced PRESENTATION_ARTEFACT segment's own
// SegmentTableSlot.slot-digest (content-addressed), never a separately stored
// inline copy. Changing the referenced slot's digest must change signed_object;
// and the SIGNATURE record carries no inline presentation digest at all, only
// sig-presentation-ref, so the slot is the only possible source.
func TestFR_064_SignedObjectSourcesPresentationDigestFromReferencedSlot(t *testing.T) {
	presRef := sigFixtureUnitID(0x30)
	slotDigestV1 := Digest{0x55}
	slotDigestV2 := Digest{0x66} // a re-rendered / changed presentation artefact

	sig := SignatureRecord{
		ParamSet:           1,
		SignedObject:       Digest{}, // not used by the assembly path
		Coverage:           CoverageDescriptor{Mode: CoverageModeTotal, Covered: []SegmentRange{{Start: 0, End: 2}}},
		CredChainRef:       sigFixtureUnitID(0x10),
		TimeAttestationRef: sigFixtureUnitID(0x20),
		PresentationRef:    presRef,
	}
	tcRoot := Digest{0x01}
	structDigest := Digest{0x02}

	resolveV1 := func(ref pdlfmt.UnitID) (Digest, bool) {
		if ref == presRef {
			return slotDigestV1, true
		}
		return Digest{}, false
	}
	resolveV2 := func(ref pdlfmt.UnitID) (Digest, bool) {
		if ref == presRef {
			return slotDigestV2, true
		}
		return Digest{}, false
	}

	so1, err := SignedObjectForSignature(sig, tcRoot, structDigest, resolveV1)
	if err != nil {
		t.Fatalf("SignedObjectForSignature v1: %v", err)
	}

	// The signed_object must equal one assembled with the slot digest directly
	// as the presentation input -- i.e. it is sourced from the slot.
	covDigest, _ := sig.Coverage.Digest()
	wantV1, _ := SignedObject(SignedObjectInput{
		TCRoot: tcRoot, StructureDigest: structDigest,
		PresentationArtefactDigest: slotDigestV1, CoverageDescriptorDigest: covDigest,
	})
	if so1 != wantV1 {
		t.Fatal("signed_object was not assembled from the referenced slot's digest")
	}

	// Changing the referenced slot's digest changes signed_object.
	so2, err := SignedObjectForSignature(sig, tcRoot, structDigest, resolveV2)
	if err != nil {
		t.Fatalf("SignedObjectForSignature v2: %v", err)
	}
	if so2 == so1 {
		t.Fatal("changing the referenced presentation slot digest did not change signed_object")
	}

	// An unresolved presentation ref is rejected (never silently substituted).
	resolveNone := func(pdlfmt.UnitID) (Digest, bool) { return Digest{}, false }
	if _, err := SignedObjectForSignature(sig, tcRoot, structDigest, resolveNone); !errors.Is(err, ErrPresentationRefUnresolved) {
		t.Errorf("unresolved presentation ref: err = %v, want ErrPresentationRefUnresolved", err)
	}
}
