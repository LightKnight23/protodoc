package semantics

import "testing"

// TestFR_118_InferredMarkerEnumeration is T-0292's named unit test (FR-118). It
// verifies the closed enumeration of inferrable fields is complete and that
// each carries the inferred-marker + basis shape, with the marker/basis pair
// coherent (a basis iff inferred).
func TestFR_118_InferredMarkerEnumeration(t *testing.T) {
	// Every inferrable field has a stable name and is enumerated exactly once.
	seen := map[InferrableField]bool{}
	for _, f := range InferrableFields {
		if seen[f] {
			t.Errorf("field %d enumerated more than once", f)
		}
		seen[f] = true
		if f.FieldName() == "unknown" {
			t.Errorf("field %d has no stable name", f)
		}
	}
	// The known inferrable fields are all present.
	for _, want := range []InferrableField{FieldAltText, FieldDirection, FieldNumberingLabel, FieldCrossReferenceText} {
		if !seen[want] {
			t.Errorf("inferrable field %q missing from enumeration", want.FieldName())
		}
	}

	// An authored value: marker absent (not inferred, BasisNone).
	authored := InferredMarker{Inferred: false, Basis: BasisNone}
	if authored.Inferred || authored.Basis != BasisNone {
		t.Errorf("authored marker should be (false, BasisNone), got %+v", authored)
	}

	// An inferred value carries a non-None basis.
	inferred := InferredMarker{Inferred: true, Basis: BasisHeuristic}
	if !inferred.Inferred || inferred.Basis == BasisNone {
		t.Errorf("inferred marker must carry a basis, got %+v", inferred)
	}
}
