// Package semantics implements Protodoc's M15 accessibility & semantic content-
// model extensions: base-writing-direction, the ROOT_SEQUENCE reading-order
// record, computed-inline isolation, the accessibility role map, table header
// scope + cell tiling, alt-text, ordered-sequence numbering, cross-reference
// presentation-function + staleness, 2-D regions, and inferred-value markers,
// plus their validator rules and the rule registry (owned by later tasks). The
// constructs are built PROVISIONALLY against the T-0267 design-ruling memo
// (clarify-002.md, OPEN awaiting Eyvar); they are reversible until that ruling.
package semantics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConformanceCase_M15_DesignRulingRecorded is T-0267's named conformance
// test (vector id ConformanceCase_M15_DesignRulingRecorded).
// It verifies the design-ruling memo (clarify-002.md) exists, enumerates all 8
// construct groups, flags the M04-only spine dependency-edge gap, and is
// HONESTLY recorded as OPEN awaiting Eyvar -- it must NOT claim a fabricated
// approval. The M15 constructs are provisional until Eyvar rules.
func TestConformanceCase_M15_DesignRulingRecorded(t *testing.T) {
	memoPath := filepath.Join("..", "..", "specs", "001-protodoc-format-core", "clarify-002.md")
	data, err := os.ReadFile(memoPath)
	if err != nil {
		t.Fatalf("read clarify-002.md: %v", err)
	}
	text := string(data)

	// Honestly OPEN: awaiting Eyvar, no fabricated sign-off.
	if !strings.Contains(text, "OPEN") || !strings.Contains(text, "awaiting Eyvar") {
		t.Error("the design-ruling memo must be recorded OPEN awaiting Eyvar (no fabricated approval)")
	}
	if !strings.Contains(text, "No Eyvar sign-off is fabricated") {
		t.Error("the memo must state explicitly that no Eyvar sign-off is fabricated")
	}
	if strings.Contains(text, "approved by Eyvar") || strings.Contains(text, "Eyvar approved") {
		t.Error("the memo must not claim Eyvar approval that has not been given")
	}

	// Enumerates all 8 construct groups (numbered 1..8).
	for i := 1; i <= 8; i++ {
		marker := string(rune('0'+i)) + "."
		if !strings.Contains(text, marker) {
			t.Errorf("memo must enumerate construct group %d", i)
		}
	}

	// Names every gated requirement. The ids are built at RUNTIME (not written
	// as bare tokens) because T-0267 is requirement-less -- a bare token here
	// would falsely register these requirements as covered by this test.
	gatedNums := []string{"032", "033", "034", "036", "037", "038", "039", "040", "082", "083", "084", "085", "099", "118"}
	for _, num := range gatedNums {
		req := "FR-" + num
		if !strings.Contains(text, req) {
			t.Errorf("memo must name gated requirement %s", req)
		}
	}

	// Flags the M04-only spine dependency-edge gap (T-0275/T-0276 retrofit
	// M08/M05).
	if !strings.Contains(text, "M04") || !strings.Contains(text, "spine dependency-edge gap") {
		t.Error("memo must flag the M04-only spine dependency-edge gap")
	}

	// Dated.
	if !strings.Contains(text, "2026-09-18") {
		t.Error("memo must carry a request date")
	}
}
