package content

import (
	"bytes"
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_084_CrossReferenceResolvesToOnePresentUnit is T-0369's named test.
// A CROSS_REFERENCE round-trips byte-exact; a target not resolving to
// exactly one present unit is rejected (FR-084 stores target identity, not
// frozen text); staleness is determined without computing layout (FR-085).
func TestFR_084_CrossReferenceResolvesToOnePresentUnit(t *testing.T) {
	xrefID, _ := testMintID()
	target, _ := testMintID()
	anchorRun, _ := testMintID()

	for _, kind := range []XrefKind{XrefInternalHyperlink, XrefCitation, XrefTableOfContentsEntry} {
		x := CrossReference{
			ID:     xrefID,
			Target: target,
			Kind:   kind,
			Anchor: AnchorPoint{RunID: anchorRun, BirthOrdinal: 2, Side: SideBefore, Boundary: BoundaryInside},
		}
		enc := EncodeCrossReference(nil, x)
		got, adv, err := DecodeCrossReference(enc)
		if err != nil {
			t.Fatalf("kind %d: DecodeCrossReference: %v", kind, err)
		}
		if adv != len(enc) || got != x {
			t.Fatalf("kind %d: cross-reference did not round-trip: got %+v want %+v", kind, got, x)
		}
		if !bytes.Equal(EncodeCrossReference(nil, got), enc) {
			t.Fatalf("kind %d: not byte-exact on re-encode", kind)
		}
	}

	x, _, _ := DecodeCrossReference(EncodeCrossReference(nil, CrossReference{
		ID: xrefID, Target: target, Kind: XrefCitation,
		Anchor: AnchorPoint{RunID: anchorRun, Side: SideBefore, Boundary: BoundaryInside},
	}))

	// Target present exactly once resolves and is not stale (no layout
	// computation involved -- pure presence check).
	present := map[pdlfmt.UnitID]int{target: 1}
	if err := x.ResolveTarget(present); err != nil {
		t.Fatalf("present target should resolve: %v", err)
	}
	if x.IsStale(present) {
		t.Fatalf("reference to a present target should not be stale")
	}

	// Dangling target (zero occurrences): rejected and stale.
	absent := map[pdlfmt.UnitID]int{}
	if err := x.ResolveTarget(absent); !errors.Is(err, ErrXrefTargetUnresolved) {
		t.Fatalf("dangling target: got %v, want ErrXrefTargetUnresolved", err)
	}
	if !x.IsStale(absent) {
		t.Fatalf("reference to an absent target should be stale (determined without layout)")
	}

	// Ambiguous target (two occurrences): rejected.
	ambiguous := map[pdlfmt.UnitID]int{target: 2}
	if err := x.ResolveTarget(ambiguous); !errors.Is(err, ErrXrefTargetUnresolved) {
		t.Fatalf("ambiguous target: got %v, want ErrXrefTargetUnresolved", err)
	}

	// A reserved xref-kind value is rejected on decode.
	enc := EncodeCrossReference(nil, x)
	enc[32] = 0x03 // kind octet at offset 16+16
	if _, _, err := DecodeCrossReference(enc); !errors.Is(err, ErrInvalidXrefKind) {
		t.Fatalf("reserved xref-kind 0x03 returned %v, want ErrInvalidXrefKind", err)
	}
	// Truncated encoding is rejected.
	if _, _, err := DecodeCrossReference(enc[:xrefFixedLen-1]); !errors.Is(err, ErrXrefTruncated) {
		t.Fatalf("truncated xref returned %v, want ErrXrefTruncated", err)
	}
}
