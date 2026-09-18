package integrity

import (
	"os"
	"strings"
	"testing"
)

// TestGOV_FR079_ClarificationRecorded is T-0199's named integration test
// (FR-079). It confirms the KNOWN GAP -- that a live (non-orphaned)
// annotation's authorship has no carrying field in the frozen data-model,
// despite document.abnf claiming it lives on the ann-body-block's runs -- is
// recorded in clarify.md as a dated OPEN clarification request for Eyvar/themis,
// rather than being silently resolved by adding a field to a frozen record.
func TestGOV_FR079_ClarificationRecorded(t *testing.T) {
	clarifyPath := "../../specs/001-protodoc-format-core/clarify.md"
	raw, err := os.ReadFile(clarifyPath)
	if err != nil {
		t.Fatalf("reading clarify.md: %v", err)
	}
	doc := string(raw)

	// The actor-identity requirement-id references are built at runtime so this
	// FR-079 task does not falsely register coverage of the publish-stripping
	// and custody-preservation requirements (owned by T-0201/T-0203).
	fr079 := "FR-" + "079"
	idx := strings.Index(doc, "RR-"+fr079)
	if idx < 0 {
		t.Fatalf("clarify.md has no RR-%s open clarification request", fr079)
	}
	section := doc[idx:]

	required := []string{
		"live-annotation authorship", // the question
		"orphan.author_ref",          // the only existing actor field
		"OPEN",                       // still awaiting a ruling
		"Eyvar",                      // who rules
		"2026-",                      // dated
	}
	for _, needle := range required {
		if !strings.Contains(section, needle) {
			t.Errorf("RR-%s clarification does not contain %q", fr079, needle)
		}
	}

	// It must remain OPEN, not carry a fabricated resolution.
	head := section
	if len(head) > 800 {
		head = head[:800]
	}
	if strings.Contains(head, "Resolution:") {
		t.Error("the FR-079 clarification must remain an OPEN request, not record a resolution")
	}
}
