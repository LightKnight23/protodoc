package semantics

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestConformance2D001RegionFields is T-0291's named conformance test (vector
// CONFORMANCE-2D-001-region-fields; FR-099). Rule PD-2D-001 rejects a region
// missing its linear-ref, whose linear-ref is not in ROOT_SEQUENCE, or a
// non-decorative region missing its text alternative.
func TestConformance2D001RegionFields(t *testing.T) {
	inSeq := pdUnitSem(0x40)
	rootSeq := []pdlfmt.UnitID{inSeq, pdUnitSem(0x41)}

	// Well-formed: linear-ref in ROOT_SEQUENCE, non-empty alt.
	good := Region2D{RegionID: pdUnitSem(0x01), LinearRef: inSeq, Alt: EmbeddedObjectAlt{AltText: "A diagram."}}
	if f := CheckRegion2DFields([]Region2D{good}, rootSeq); len(f) != 0 {
		t.Errorf("well-formed region should pass, got %+v", f)
	}

	// Missing linear-ref (zero unit-id).
	noRef := Region2D{RegionID: pdUnitSem(0x02), Alt: EmbeddedObjectAlt{AltText: "x"}}
	if f := CheckRegion2DFields([]Region2D{noRef}, rootSeq); len(f) != 1 || f[0].Kind != Region2DMissingLinearRef {
		t.Errorf("missing linear-ref: findings = %+v, want one MissingLinearRef", f)
	}

	// Linear-ref not in ROOT_SEQUENCE.
	badRef := Region2D{RegionID: pdUnitSem(0x03), LinearRef: pdUnitSem(0x99), Alt: EmbeddedObjectAlt{AltText: "x"}}
	f := CheckRegion2DFields([]Region2D{badRef}, rootSeq)
	if len(f) != 1 || f[0].Kind != Region2DUnresolvedLinearRef || f[0].Rule != RuleRegion2DFields {
		t.Errorf("unresolved linear-ref: findings = %+v, want one UnresolvedLinearRef", f)
	}

	// Non-decorative region missing alt text.
	noAlt := Region2D{RegionID: pdUnitSem(0x04), LinearRef: inSeq, Alt: EmbeddedObjectAlt{Decorative: false, AltText: "  "}}
	fa := CheckRegion2DFields([]Region2D{noAlt}, rootSeq)
	if len(fa) != 1 || fa[0].Kind != Region2DMissingAltText {
		t.Errorf("missing alt: findings = %+v, want one MissingAltText", fa)
	}

	// Decorative region needs no alt text.
	decor := Region2D{RegionID: pdUnitSem(0x05), LinearRef: inSeq, Alt: EmbeddedObjectAlt{Decorative: true}}
	if f := CheckRegion2DFields([]Region2D{decor}, rootSeq); len(f) != 0 {
		t.Errorf("decorative region should pass, got %+v", f)
	}
}
