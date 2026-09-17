package ledger

import (
	"os"
	"strings"
	"testing"
)

// designNotePath is the documented location of the TR-010 conditional-write
// design note, relative to this package directory (pkg/ledger), two levels
// up to the module root, matching the relative-path convention this module's
// other checked-in-artifact tests use.
const designNotePath = "../../docs/ledger-conditional-write-design-note.md"

// requiredDesignNoteSections are the two section headers T-0049's DoD
// requires the note to contain. The doc-lint fails if either the file or
// either section is missing.
var requiredDesignNoteSections = []string{
	"FR-117 vs TR-010 distinction",
	"ConditionalWriter contract",
}

// TestTR_010_DesignNoteContainsRequiredSections is T-0049's named test: a
// doc-lint asserting the conditional-write design note exists at its
// documented path and contains both required section headers, distinguishing
// FR-117 (file-internal torn-write detection) from TR-010 (external
// shared-storage last-writer-wins avoidance) and documenting the
// ConditionalWriter contract for M18's CLI implementers. It fails if the
// note file or either section is missing.
func TestTR_010_DesignNoteContainsRequiredSections(t *testing.T) {
	raw, err := os.ReadFile(designNotePath)
	if err != nil {
		t.Fatalf("TR-010 design note missing at %s: %v", designNotePath, err)
	}
	text := string(raw)

	for _, section := range requiredDesignNoteSections {
		// Require the section as a markdown header ("## <section>"), not
		// merely a passing mention, so the note is actually structured
		// around both topics.
		header := "## " + section
		if !strings.Contains(text, header) {
			t.Errorf("TR-010 design note %s is missing required section header %q", designNotePath, header)
		}
	}

	// The note must also actually reference the ConditionalWriter type and
	// the two mechanisms it distinguishes, so the headers are not empty.
	for _, needle := range []string{"ConditionalWriter", "commit ring", "conditional"} {
		if !strings.Contains(text, needle) {
			t.Errorf("TR-010 design note %s does not mention %q", designNotePath, needle)
		}
	}
}
