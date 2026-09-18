package semantics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFR_032_DirectionFieldInDataModel is T-0268's named unit test (FR-032).
// Every text-container entity's field table in data-model.md lists a
// `direction` field with a closed value set {LTR, RTL} and a 'mandatory'
// constraint (provisional per the T-0267 ruling).
func TestFR_032_DirectionFieldInDataModel(t *testing.T) {
	path := filepath.Join("..", "..", "specs", "001-protodoc-format-core", "data-model.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read data-model.md: %v", err)
	}
	text := string(data)

	// Every text-container entity section must carry a direction field row.
	// We check the three entity sections named in the T-0267 ruling.
	sections := map[string]string{
		"TextBlock": "### 2.7 TextBlock",
		"Table":     "### 2.24 Table",
		"Document":  "### 2.3 Frontmatter", // top-level document unit (document_metadata)
	}
	for name, header := range sections {
		idx := strings.Index(text, header)
		if idx < 0 {
			t.Fatalf("%s: section %q not found", name, header)
		}
		// Bound the section at the next "### 2." header.
		rest := text[idx+len(header):]
		if end := strings.Index(rest, "### 2."); end >= 0 {
			rest = rest[:end]
		}
		if !strings.Contains(rest, "direction") {
			t.Errorf("%s: field table must list a direction field", name)
		}
		if !strings.Contains(rest, "mandatory") {
			t.Errorf("%s: direction field must be mandatory", name)
		}
		if !strings.Contains(rest, "LTR") || !strings.Contains(rest, "RTL") {
			t.Errorf("%s: direction must be a closed value set {LTR, RTL}", name)
		}
	}
}
