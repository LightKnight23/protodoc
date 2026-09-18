package semantics

import "testing"

// TestConformanceINFER001MarkerConsistency is T-0293's named conformance test
// (vector CONFORMANCE-INFER-001-marker-consistency; FR-118). Rule PD-INFER-001
// requires the inferred-marker + basis pair be present when a value was not
// authored, and absent when it was.
func TestConformanceINFER001MarkerConsistency(t *testing.T) {
	// Consistent: authored value with {inferred=0, none}; inferred value with
	// {inferred=1, heuristic}.
	ok := []FieldValue{
		{Field: FieldDirection, Marker: InferredMarker{Inferred: false, Basis: BasisNone}, WasAuthored: true},
		{Field: FieldAltText, Marker: InferredMarker{Inferred: true, Basis: BasisModel}, WasAuthored: false},
	}
	if f := CheckInferMarker(ok); len(f) != 0 {
		t.Errorf("consistent markers should pass, got %+v", f)
	}

	// Authored value wrongly marked inferred -> rejected.
	badAuthored := []FieldValue{{Field: FieldDirection, Marker: InferredMarker{Inferred: true, Basis: BasisHeuristic}, WasAuthored: true}}
	if f := CheckInferMarker(badAuthored); len(f) != 1 || f[0].Rule != RuleInferMarker {
		t.Errorf("authored+inferred-marker: findings = %+v, want one PD-INFER-001", f)
	}

	// Inferred value not marked -> rejected.
	unmarked := []FieldValue{{Field: FieldAltText, Marker: InferredMarker{Inferred: false, Basis: BasisNone}, WasAuthored: false}}
	if f := CheckInferMarker(unmarked); len(f) != 1 {
		t.Errorf("inferred-but-unmarked: findings = %+v, want one rejection", f)
	}

	// Inferred value marked but no basis -> rejected.
	noBasis := []FieldValue{{Field: FieldNumberingLabel, Marker: InferredMarker{Inferred: true, Basis: BasisNone}, WasAuthored: false}}
	if f := CheckInferMarker(noBasis); len(f) != 1 {
		t.Errorf("inferred-no-basis: findings = %+v, want one rejection", f)
	}
}
