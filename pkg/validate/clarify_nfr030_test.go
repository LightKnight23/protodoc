package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNFR_030_ClarifyRulingRequestRecorded is T-0361's named unit test. M07's
// NFR-030 memory-floor reading (the floor-exempted max(ArenaFloor, 4*inputLen)
// bound adopted in T-0121, disclosed in plan.md Section 9 Conflict 3) deviates
// from NFR-030's literal text. Unlike the resolved phase-5 rulings, no task in
// M07 had recorded an OPEN ruling request for it, leaving governance tracking
// of this deviation inconsistent with the sibling disclosed-conflict pattern.
// This test verifies clarify.md now carries a matching open-ruling-request
// entry: it names NFR-030, marks the request OPEN (not a fabricated
// resolution), states the adopted floor-exempted reading, is dated, and
// cross-references plan.md Section 9 Conflict 3 and T-0121.
func TestNFR_030_ClarifyRulingRequestRecorded(t *testing.T) {
	clarifyPath := filepath.Join("..", "..", "specs", "001-protodoc-format-core", "clarify.md")
	data, err := os.ReadFile(clarifyPath)
	if err != nil {
		t.Fatalf("read clarify.md: %v", err)
	}
	text := string(data)

	// Isolate the open-ruling-requests section so a stray mention elsewhere
	// cannot satisfy the assertions.
	const sectionHeader = "## Open ruling requests"
	idx := strings.Index(text, sectionHeader)
	if idx < 0 {
		t.Fatalf("clarify.md has no %q section recording open disclosed-conflict ruling requests", sectionHeader)
	}
	section := text[idx:]

	required := []struct {
		desc   string
		needle string
	}{
		{"names NFR-030", "NFR-030"},
		{"marks the request OPEN (not resolved)", "OPEN"},
		{"states the floor-exempted adopted reading", "max(ArenaFloor, 4 * inputLen)"},
		{"cross-references plan.md Section 9 Conflict 3", "Section 9 Conflict 3"},
		{"cross-references the T-0121 arena task", "T-0121"},
		{"is dated", "2026-09-15"},
	}
	for _, r := range required {
		if !strings.Contains(section, r.needle) {
			t.Errorf("NFR-030 open ruling request does not %s (missing %q)", r.desc, r.needle)
		}
	}

	// Honesty guard: the entry must NOT claim a resolution/ruling was made.
	// It records an OPEN request only; a fabricated "Eyvar ruled" line here
	// would be a false sign-off.
	entryStart := strings.Index(section, "RR-NFR-030: ")
	if entryStart >= 0 {
		entry := section[entryStart:]
		if strings.Contains(entry, "Eyvar ruled") || strings.Contains(entry, "Resolution:") {
			t.Error("NFR-030 ruling request must remain OPEN; it must not record a resolution or a ruling that has not been made")
		}
	}
}
