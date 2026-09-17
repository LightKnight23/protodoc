package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCON_021_RegistryTurnaroundSLAPublished is T-0061's named unit test.
// CON-021 is a governance/process obligation: publish a maximum extension-
// registry review turnaround in business days and track observed turnaround
// against it. This is a documentation-completeness check (not a wire-format
// fixture): it verifies docs/token-registry-governance.md exists, names a
// business-day SLA figure and a tracking mechanism, and is cross-referenced
// from the contracts README's registry section.
func TestCON_021_RegistryTurnaroundSLAPublished(t *testing.T) {
	root := filepath.Join("..", "..")
	govPath := filepath.Join(root, "docs", "token-registry-governance.md")
	data, err := os.ReadFile(govPath)
	if err != nil {
		t.Fatalf("token-registry governance doc missing at %s: %v", govPath, err)
	}
	doc := string(data)

	required := []struct {
		desc   string
		needle string
	}{
		{"names CON-021", "CON-021"},
		{"scopes the registered tier", "registered-tier"},
		{"publishes a maximum turnaround", "maximum review turnaround"},
		{"states the SLA in business days", "business days"},
		{"gives a concrete SLA figure", "15 business days"},
		{"defines a per-request tracking mechanism", "per request"},
		{"defines per-release reporting", "per release"},
		{"names the governance metric", "G-REG"},
	}
	for _, r := range required {
		if !strings.Contains(doc, r.needle) {
			t.Errorf("governance doc does not %s (missing %q)", r.desc, r.needle)
		}
	}

	// The contracts README's registry section must cross-reference the doc.
	readmePath := filepath.Join(root, "specs", "001-protodoc-format-core", "contracts", "README.md")
	rdata, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("contracts README missing: %v", err)
	}
	readme := string(rdata)
	if !strings.Contains(readme, "token-registry-governance.md") {
		t.Error("contracts README does not cross-reference docs/token-registry-governance.md")
	}
	if !strings.Contains(readme, "CON-021") {
		t.Error("contracts README registry section does not cite CON-021")
	}
}
