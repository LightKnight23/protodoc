package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCON_017_AnalyzeInputNotePresent is T-0360's named unit test. T-0126's
// A-CLASS lint guard is independently CI-checkable, but the accompanying
// clause "a note added to phase-5 analyze inputs recording that CON-017 is
// satisfied by the absence of a second writer conformance class" is a separate
// documentation deliverable the lint cannot verify. This test verifies that
// deliverable: the note file exists under the phase-5 analyze-inputs path,
// records the satisfied-by-absence disposition, and cites T-0126's lint as the
// CI-checkable evidence, so the analyze phase can consume it without further
// clarification.
func TestCON_017_AnalyzeInputNotePresent(t *testing.T) {
	notePath := filepath.Join("..", "..", "specs", "001-protodoc-format-core",
		"analyze-inputs", "con-017-single-writer-class.md")

	data, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("CON-017 analyze-input note missing at %s: %v", notePath, err)
	}
	text := string(data)

	// Required content: the disposition (satisfied by absence of a second
	// writer class) and the T-0126 lint citation.
	requiredPhrases := []struct {
		desc   string
		needle string
	}{
		{"names CON-017", "CON-017"},
		{"states satisfied by absence", "absence"},
		{"names the single writer conformance class concept", "writer conformance class"},
		{"cites the T-0126 evidence task", "T-0126"},
		{"cites the A-CLASS lint guard test", "TestCON_017_SingleWriterConformanceClassAsserted"},
	}
	for _, p := range requiredPhrases {
		if !strings.Contains(text, p.needle) {
			t.Errorf("CON-017 analyze-input note does not %s (missing %q)", p.desc, p.needle)
		}
	}
}
