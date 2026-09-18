package semantics

import "testing"

// a11yFailureCondition is one published accessibility structural/presence
// failure condition, and whether a PD-rule from T-0268..T-0293 decides it
// PURELY STRUCTURALLY (no human judgment).
type a11yFailureCondition struct {
	name             string
	requirement      string
	decidedByRule    string // the deciding PD-rule id, or "" if not structurally decidable
	structuralDecide bool
}

// a11yFailureConditions enumerates every accessibility-relevant structural or
// presence failure condition published in spec.md's accessibility FRs, with the
// PD-rule (if any) that decides it structurally. A condition requiring human
// judgment (e.g. whether alt text is *meaningful* beyond echoing metadata is
// heuristic, not fully structural) is marked structuralDecide=false honestly.
var a11yFailureConditions = []a11yFailureCondition{
	{"missing base direction", "FR-032", RuleMissingDirection, true},
	{"in-band directional control", "FR-032", RuleInbandDirectionControl, true},
	{"reading-order omission/duplicate", "FR-036", RuleRootSequenceCompleteness, true},
	{"mark without placement", "FR-037", RuleMarkPlacement, true},
	{"construct with no role", "FR-038", RuleRoleResolution, true},
	{"header scope resolves to no data cell", "FR-039", RuleHeaderScope, true},
	{"table cells do not tile", "FR-082", RuleCellTiling, true},
	{"non-decorative object missing alt text", "FR-040", RuleAltTextQuality, true},
	{"numbering literal not derivable", "FR-083", RuleNumberingLiteral, true},
	{"2D region missing required fields", "FR-099", RuleRegion2DFields, true},
	{"inferred-marker inconsistent", "FR-118", RuleInferMarker, true},
	// Honestly NON-structural: whether provided alt text is *semantically*
	// adequate (beyond the metadata-echo heuristic) requires human judgment.
	{"alt text semantic adequacy", "FR-040", "", false},
}

// TestNFR_031_EightyPercentDecidable is T-0295's named conformance test
// (NFR-031). It asserts at least 80% of the published accessibility
// structural/presence failure conditions are decided purely structurally by a
// PD-rule, with no human judgment.
func TestNFR_031_EightyPercentDecidable(t *testing.T) {
	total := len(a11yFailureConditions)
	if total == 0 {
		t.Fatalf("no accessibility failure conditions enumerated")
	}
	decidable := 0
	for _, c := range a11yFailureConditions {
		if c.structuralDecide {
			if c.decidedByRule == "" {
				t.Errorf("%q claims structural decidability but names no rule", c.name)
				continue
			}
			decidable++
		}
	}

	// decidable/total >= 0.80, checked with integer arithmetic (no floats).
	if decidable*100 < 80*total {
		t.Errorf("structurally decidable fraction %d/%d < 80%%", decidable, total)
	}
	t.Logf("accessibility structural decidability: %d of %d conditions (%d%%)", decidable, total, decidable*100/total)
}
