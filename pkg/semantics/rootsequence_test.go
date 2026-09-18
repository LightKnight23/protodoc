package semantics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_036_RootSequenceSchema is T-0273's named unit test (FR-036).
// data-model.md defines ROOT_SEQUENCE with an ordered list of unit references,
// one entry per content unit, stated to be the authoritative reading order
// independent of storage/append order; the Go schema mirrors it.
func TestFR_036_RootSequenceSchema(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "specs", "001-protodoc-format-core", "data-model.md"))
	if err != nil {
		t.Fatalf("read data-model.md: %v", err)
	}
	idx := strings.Index(string(data), "### 2.27 ROOT_SEQUENCE")
	if idx < 0 {
		t.Fatal("data-model.md must define the ROOT_SEQUENCE entity (2.27)")
	}
	section := string(data)[idx:]
	if end := strings.Index(section, "## 3. Relationships"); end >= 0 {
		section = section[:end]
	}
	for _, need := range []string{
		"rs_order", "one entry per content unit", "authoritative",
	} {
		if !strings.Contains(section, need) {
			t.Errorf("ROOT_SEQUENCE section must state %q", need)
		}
	}
	if !strings.Contains(strings.ToLower(section), "independent of storage") {
		t.Error("ROOT_SEQUENCE section must state it is independent of storage order")
	}

	// The Go schema mirrors it: ordered unit list with position lookup.
	seq := RootSequence{RSID: pdUnitSem(0x01), Order: unitList(0x10, 0x20, 0x30)}
	if seq.Len() != 3 {
		t.Fatalf("Len = %d, want 3", seq.Len())
	}
	if p, ok := seq.Position(pdUnitSem(0x20)); !ok || p != 1 {
		t.Errorf("Position(0x20) = %d,%v, want 1,true", p, ok)
	}
	if _, ok := seq.Position(pdUnitSem(0x99)); ok {
		t.Error("Position of an absent unit must report not-present")
	}
}

// unitList builds a slice of unit ids from seed bytes.
func unitList(seeds ...byte) []pdlfmt.UnitID {
	out := make([]pdlfmt.UnitID, len(seeds))
	for i, s := range seeds {
		out[i] = pdUnitSem(s)
	}
	return out
}
