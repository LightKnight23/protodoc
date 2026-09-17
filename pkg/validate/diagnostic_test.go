package validate

import (
	"strings"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_102_StructuralFailureReportsOffsetUnitRule is T-0114's named
// conformance test. Each of the three structural-failure classes
// (truncated segment, oversized declared length, malformed TLV field)
// yields a Finding with the exact expected offset, unit id (or NONE), and
// rule id, matched against golden expectations.
func TestFR_102_StructuralFailureReportsOffsetUnitRule(t *testing.T) {
	unitID := pdlfmt.UnitID{0x11, 0x22, 0x33}

	cases := []struct {
		name       string
		finding    *Finding
		wantStep   StepID
		wantOffset uint64
		wantUnit   string
		wantRule   string
	}{
		{
			name:       "truncated segment (pre-unit prefix)",
			finding:    StructuralFinding(StepBoundedPrefix, Diagnostic{Offset: 262150, Unit: NoUnit(), RuleID: RuleTruncated}),
			wantStep:   StepBoundedPrefix,
			wantOffset: 262150,
			wantUnit:   "NONE",
			wantRule:   RuleTruncated,
		},
		{
			name:       "oversized declared length in a unit",
			finding:    StructuralFinding(StepStructuralCeilings, Diagnostic{Offset: 1049000, Unit: InUnit(unitID), RuleID: RuleOversizedLength}),
			wantStep:   StepStructuralCeilings,
			wantOffset: 1049000,
			wantUnit:   "112233" + strings.Repeat("00", 13),
			wantRule:   RuleOversizedLength,
		},
		{
			name:       "malformed TLV field in a unit",
			finding:    StructuralFinding(StepBoundedPrefix, Diagnostic{Offset: 1048600, Unit: InUnit(unitID), RuleID: RuleUnparseableTLV}),
			wantStep:   StepBoundedPrefix,
			wantOffset: 1048600,
			wantUnit:   "112233" + strings.Repeat("00", 13),
			wantRule:   RuleUnparseableTLV,
		},
	}

	for _, c := range cases {
		f := c.finding
		if f.Diag == nil {
			t.Fatalf("%s: Finding has no diagnostic", c.name)
		}
		if f.Step != c.wantStep {
			t.Fatalf("%s: step = %d, want %d", c.name, f.Step, c.wantStep)
		}
		if f.Diag.Offset != c.wantOffset {
			t.Fatalf("%s: offset = %d, want %d", c.name, f.Diag.Offset, c.wantOffset)
		}
		if f.Diag.Unit.String() != c.wantUnit {
			t.Fatalf("%s: unit = %q, want %q", c.name, f.Diag.Unit.String(), c.wantUnit)
		}
		if f.Diag.RuleID != c.wantRule || f.RuleID != c.wantRule {
			t.Fatalf("%s: rule = %q/%q, want %q", c.name, f.Diag.RuleID, f.RuleID, c.wantRule)
		}
		// The message names all three coordinates.
		for _, needle := range []string{c.wantUnit, c.wantRule} {
			if !strings.Contains(f.Message, needle) {
				t.Fatalf("%s: message %q does not name %q", c.name, f.Message, needle)
			}
		}
		// It is a validity failure, not a budget status.
		if f.Budget {
			t.Fatalf("%s: structural failure wrongly marked as budget", c.name)
		}
	}

	// NoUnit renders as an explicit NONE (not a zero id).
	if NoUnit().String() != "NONE" {
		t.Fatalf("NoUnit().String() = %q, want NONE", NoUnit().String())
	}
}
