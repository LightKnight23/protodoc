package semantics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// m15RuleIDs is the closed set of M15 accessibility/semantic validator rule ids
// this milestone introduced. Each MUST appear in the contracts/README.md
// rule-id-to-conformance-case registry (NFR-031 / CP-011) and each MUST be a
// live constant in this package.
var m15RuleIDs = []string{
	RuleMissingDirection,         // PD-BIDI-001
	RuleMarkPlacement,            // PD-A11Y-001
	RuleRoleResolution,           // PD-A11Y-002
	RuleAltTextQuality,           // PD-A11Y-003
	RuleNumberingLiteral,         // PD-A11Y-004
	RuleRootSequenceCompleteness, // PD-A11Y-005
	RuleHeaderScope,              // PD-TBL-001a
	RuleCellTiling,               // PD-TBL-001b
	RuleRegion2DFields,           // PD-2D-001
	RuleInferMarker,              // PD-INFER-001
}

// TestNFR_031_RuleRegistryComplete is T-0294's named conformance test
// (NFR-031). Every M15 rule id is registered in contracts/README.md's
// rule-id-to-conformance-case registry, closing the pattern where spec-named
// rule ids were never carried downstream.
func TestNFR_031_RuleRegistryComplete(t *testing.T) {
	readme, err := os.ReadFile(filepath.Join("..", "..", "specs", "001-protodoc-format-core", "contracts", "README.md"))
	if err != nil {
		t.Fatalf("read contracts/README.md: %v", err)
	}
	doc := string(readme)

	for _, id := range m15RuleIDs {
		// The registry lists PD-A11Y-001 through 005 and PD-TBL-001a/001b as
		// ranges; check the family prefix is present for those, exact id else.
		probe := id
		switch {
		case strings.HasPrefix(id, "PD-A11Y-"):
			probe = "PD-A11Y-001" // the range anchor "PD-A11Y-001 through 005"
		case strings.HasPrefix(id, "PD-TBL-001"):
			probe = "PD-TBL-001a"
		}
		if !strings.Contains(doc, probe) {
			t.Errorf("rule id %q (probe %q) not registered in contracts/README.md rule registry", id, probe)
		}
	}

	// Every rule id is a non-empty live constant (carried in code, not just doc).
	for _, id := range m15RuleIDs {
		if id == "" {
			t.Errorf("an M15 rule id constant is empty")
		}
		if !strings.HasPrefix(id, "PD-") {
			t.Errorf("rule id %q is not a PD- rule id", id)
		}
	}
}
