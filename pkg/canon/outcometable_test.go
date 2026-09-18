package canon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFR_124_EmptyStateOutcomeTableSchemaComplete is T-0317's named unit test
// (FR-124). The empty-state outcome table (in the contracts set) must name the
// exact expected output/verdict for the BOTTOM fixture for every one of the
// five reader roles; the schema check fails if any role is missing.
func TestFR_124_EmptyStateOutcomeTableSchemaComplete(t *testing.T) {
	path := filepath.Join("..", "..", "specs", "001-protodoc-format-core", "contracts", "empty-state-outcome-table.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read outcome table: %v", err)
	}
	doc := strings.ToLower(string(raw))

	roles := []string{"extraction", "preview", "reflow", "rasterisation", "validation"}
	for _, role := range roles {
		// Each role must appear as a table row (a line beginning "| <role>").
		found := false
		for _, line := range strings.Split(doc, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "| "+role) {
				// The row must carry a non-empty outcome cell beyond the role.
				cells := strings.Split(trimmed, "|")
				if len(cells) >= 3 && strings.TrimSpace(cells[2]) != "" {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("outcome table missing a defined-outcome row for reader role %q", role)
		}
	}
}
