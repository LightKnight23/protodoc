package cli

import (
	"strings"
	"testing"
)

// TestTR_001_FindingCarriesAllFiveFields is T-0328's named conformance test
// (TR-001). One fixture per validator rule category confirms all 5 fields are
// non-empty and the normative-statement id matches the checked-in map; a
// finding missing any field is a build failure.
func TestTR_001_FindingCarriesAllFiveFields(t *testing.T) {
	// One representative rule per category present in the checked-in map.
	// Gap-owned rule ids (DISC, NFC, DUP) are built at runtime so this test
	// does not falsely register their conformance case.
	fixtures := []struct {
		rule    string
		wantReq string
	}{
		{"PD-VARINT-001", "FR-001"},
		{"PD-SORT-001", "NFR-002"},
		{pd("DISC", 1), "FR-013"},
		{pd("NFC", 1), "FR-020"},
		{"PD-CEILING-001", "NFR-030"},
		{pd("DUP", 1), "FR-110"},
		{"PD-A11Y-005", "FR-036"},
	}

	for _, fx := range fixtures {
		f, ok := NewFiveFieldFinding(fx.rule, SeverityError, "0102030405060708090a0b0c0d0e0f10", 4096)
		if !ok {
			t.Errorf("rule %s not in the checked-in rule->requirement map", fx.rule)
			continue
		}
		// All five fields non-empty / present.
		if f.RuleID == "" {
			t.Errorf("%s: rule_id empty", fx.rule)
		}
		if f.Severity == "" {
			t.Errorf("%s: severity empty", fx.rule)
		}
		if f.UnitID == "" {
			t.Errorf("%s: unit_id empty", fx.rule)
		}
		if f.OctetOffset == 0 {
			t.Errorf("%s: octet_offset zero (fixture used 4096)", fx.rule)
		}
		if f.RequirementID == "" {
			t.Errorf("%s: requirement_id empty", fx.rule)
		}
		// The requirement id matches the checked-in map.
		if f.RequirementID != fx.wantReq {
			t.Errorf("%s: requirement_id = %s, want %s", fx.rule, f.RequirementID, fx.wantReq)
		}
		// Requirement ids are well-formed normative-statement ids.
		if !strings.HasPrefix(f.RequirementID, "FR-") && !strings.HasPrefix(f.RequirementID, "NFR-") &&
			!strings.HasPrefix(f.RequirementID, "CON-") && !strings.HasPrefix(f.RequirementID, "TR-") {
			t.Errorf("%s: requirement_id %q is not a normative-statement id", fx.rule, f.RequirementID)
		}
	}

	// An unmapped rule id is refused (build-failing at the call site).
	if _, ok := NewFiveFieldFinding("PD-UNKNOWN-999", SeverityError, "x", 1); ok {
		t.Errorf("an unmapped rule id must not produce a finding")
	}
}
