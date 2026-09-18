package semantics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConformanceA11Y005RootSequenceCompleteness is T-0274's named conformance
// test (vector id CONFORMANCE-A11Y-005-rootsequence-completeness; FR-036). The
// ROOT_SEQUENCE production is in document.abnf, and rule PD-A11Y-005 rejects a
// reading order that omits a unit or lists one twice, naming the unit id; a
// complete reading order passes.
func TestConformanceA11Y005RootSequenceCompleteness(t *testing.T) {
	// document.abnf carries the ROOT_SEQUENCE production.
	abnf, err := os.ReadFile(filepath.Join("..", "..", "specs", "001-protodoc-format-core", "contracts", "document.abnf"))
	if err != nil {
		t.Fatalf("read document.abnf: %v", err)
	}
	for _, tok := range []string{"root-sequence", "rs-order", "rs-discriminant"} {
		if !strings.Contains(string(abnf), tok) {
			t.Errorf("document.abnf must define %s", tok)
		}
	}

	content := unitList(0x10, 0x20, 0x30)

	// Complete reading order -> passes.
	complete := RootSequence{RSID: pdUnitSem(0x01), Order: unitList(0x10, 0x20, 0x30)}
	if f := complete.CheckRootSequenceCompleteness(content); len(f) != 0 {
		t.Errorf("complete reading order should pass, got %+v", f)
	}

	// Omission (0x30 missing) -> rejected naming 0x30.
	omit := RootSequence{RSID: pdUnitSem(0x01), Order: unitList(0x10, 0x20)}
	f := omit.CheckRootSequenceCompleteness(content)
	if len(f) != 1 || f[0].Kind != ViolationOmission || f[0].UnitID != pdUnitSem(0x30) {
		t.Errorf("omission: findings = %+v, want one omission naming 0x30", f)
	}
	if f[0].Rule != RuleRootSequenceCompleteness {
		t.Errorf("rule = %q, want %q", f[0].Rule, RuleRootSequenceCompleteness)
	}

	// Duplicate (0x20 twice) -> rejected naming 0x20.
	dup := RootSequence{RSID: pdUnitSem(0x01), Order: unitList(0x10, 0x20, 0x20, 0x30)}
	fd := dup.CheckRootSequenceCompleteness(content)
	foundDup := false
	for _, x := range fd {
		if x.Kind == ViolationDuplicate && x.UnitID == pdUnitSem(0x20) {
			foundDup = true
		}
	}
	if !foundDup {
		t.Errorf("duplicate: findings = %+v, want a duplicate naming 0x20", fd)
	}
}
