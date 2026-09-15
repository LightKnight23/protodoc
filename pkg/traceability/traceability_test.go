package traceability

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestNFR_029_TraceabilityCheckerParsesAllRuleIDs is T-0352's named test.
// It proves the checker's two halves independently:
//
//  1. ExtractRequirementIDs and ExtractRuleIDs correctly parse every
//     shape of id this repository actually uses, including the
//     alphanumeric-segment rule ids (A11Y, 2D) that a naive "letters only"
//     pattern would miss, and never conflate a requirement id with a rule
//     id or vice versa.
//  2. A synthetic fixture with one deliberately-orphaned rule id (present
//     in the "spec"/"contract" text, absent from the "corpus" source
//     tree) is correctly reported as a gap, while a sibling rule id that
//     does appear in the corpus is not -- proving FindGaps fails closed
//     rather than accepting or rejecting everything indiscriminately.
func TestNFR_029_TraceabilityCheckerParsesAllRuleIDs(t *testing.T) {
	const specSnippet = `
**FR-042** *(ubiquitous, must)*

> Some requirement text mentioning CON-007 and NFR-013 in passing.

Validator rule PD-RING-001 rejects this. Validator rule PD-A11Y-002 and
PD-2D-001 also apply. TR-006 governs the fixed prefix.
`
	gotReq := ExtractRequirementIDs(specSnippet)
	wantReq := []string{"CON-007", "FR-042", "NFR-013", "TR-006"}
	if !reflect.DeepEqual(gotReq, wantReq) {
		t.Errorf("ExtractRequirementIDs(specSnippet) = %v, want %v", gotReq, wantReq)
	}

	gotRule := ExtractRuleIDs(specSnippet)
	wantRule := []string{"PD-2D-001", "PD-A11Y-002", "PD-RING-001"}
	if !reflect.DeepEqual(gotRule, wantRule) {
		t.Errorf("ExtractRuleIDs(specSnippet) = %v, want %v", gotRule, wantRule)
	}

	// Cross-contamination check: a rule-id pattern must never also be
	// reported as a requirement id (they share no prefix, but the two
	// regexes are maintained side by side and could drift onto
	// overlapping matches if edited carelessly).
	for _, id := range gotRule {
		for _, other := range gotReq {
			if id == other {
				t.Errorf("id %q reported as both a requirement id and a rule id", id)
			}
		}
	}

	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.md")
	contractPath := filepath.Join(dir, "contract.abnf")
	corpusDir := filepath.Join(dir, "corpus")

	// The "spec"/"contract" text names two rule ids: one that the
	// "corpus" source tree below actually exercises, and one --
	// PD-ORPHAN-001 -- that it deliberately does not.
	if err := os.WriteFile(specPath, []byte("Validator rule PD-COVERED-001 and PD-ORPHAN-001 both apply. FR-001 too.\n"), 0o644); err != nil {
		t.Fatalf("writing synthetic spec fixture: %v", err)
	}
	if err := os.WriteFile(contractPath, []byte("; PD-COVERED-001 and PD-ORPHAN-001 restated here, contracts-side.\n"), 0o644); err != nil {
		t.Fatalf("writing synthetic contract fixture: %v", err)
	}
	if err := os.MkdirAll(corpusDir, 0o755); err != nil {
		t.Fatalf("making synthetic corpus dir: %v", err)
	}
	const corpusSrc = `package fixture

// Implements: FR-001.
// This test exercises rule PD-COVERED-001.
func exampleCoveredCase() {}
`
	if err := os.WriteFile(filepath.Join(corpusDir, "covered_test.go"), []byte(corpusSrc), 0o644); err != nil {
		t.Fatalf("writing synthetic corpus fixture: %v", err)
	}

	report, err := Audit(specPath, []string{contractPath}, corpusDir)
	if err != nil {
		t.Fatalf("Audit(synthetic fixture) = _, %v, want nil error", err)
	}

	if len(report.Gaps) != 1 {
		t.Fatalf("Audit(synthetic fixture).Gaps = %v, want exactly one gap", report.Gaps)
	}
	got := report.Gaps[0]
	if got.ID != "PD-ORPHAN-001" || got.Kind != "rule" {
		t.Errorf("Audit(synthetic fixture).Gaps[0] = %+v, want {ID: PD-ORPHAN-001, Kind: rule}", got)
	}
	if report.AllCovered() {
		t.Error("Audit(synthetic fixture).AllCovered() = true, want false: PD-ORPHAN-001 has zero matching cases")
	}

	// FR-001 and PD-COVERED-001 must NOT appear as gaps: both occur
	// literally in covered_test.go.
	for _, g := range report.Gaps {
		if g.ID == "FR-001" || g.ID == "PD-COVERED-001" {
			t.Errorf("Audit(synthetic fixture) incorrectly reported covered id %q as a gap", g.ID)
		}
	}
}
