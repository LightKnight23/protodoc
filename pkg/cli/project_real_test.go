package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestTR_012_ProjectVerbReadsRealFile is T-0376's named integration test
// (TR-012/TR-004, DEFECT-2026-09-19 fix). The production project backend
// (realProjectStateFor, wired by init) reads real files: an absent/garbage file
// fails the CP-006 gate; a real document with CONTENT segments produces a
// non-empty, deterministic projection derived from the actual frame bytes.
func TestTR_012_ProjectVerbReadsRealFile(t *testing.T) {
	// (1) Absent file -> INVALID (CP-006 validate-first).
	res := runProject([]string{filepath.Join(t.TempDir(), "absent.pdl")}, nil)
	if res.Status != "INVALID" {
		t.Errorf("absent file: status=%s, want INVALID", res.Status)
	}

	// (2) A real document with CONTENT frames -> non-empty projection keyed by
	// the AUTHORED unit-id.
	doc, ids := writeDocWithRealFrames(t, cliUnit(0xA1), cliUnit(0xB2))
	res = runProject([]string{doc, "--format=text"}, nil)
	if res.Status != "OK" {
		t.Fatalf("content doc: status=%s (%+v), want OK", res.Status, res.Findings)
	}
	projection := res.Extra["projection"].(string)
	if strings.TrimSpace(projection) == "" {
		t.Errorf("projection of a document with content must be non-empty")
	}
	// The projection is keyed by the authored unit-id (hex of ids[0] appears).
	if !strings.Contains(projection, "a1") {
		t.Errorf("projection should include the authored unit-id (a1..), got %q", projection)
	}
	_ = ids

	// Deterministic: projecting twice gives identical output.
	res2 := runProject([]string{doc, "--format=text"}, nil)
	if res2.Extra["projection"].(string) != projection {
		t.Errorf("project output not deterministic across runs")
	}

	// html format also works over the real content.
	resH := runProject([]string{doc, "--format=html"}, nil)
	if resH.Status != "OK" || !strings.Contains(resH.Extra["projection"].(string), "data-unit") {
		t.Errorf("html projection over real content failed: %+v", resH.Extra)
	}

	// (3) Empty doc -> empty projection, still OK.
	empty := writeValidPrefix(t)
	res = runProject([]string{empty, "--format=text"}, nil)
	if res.Status != "OK" || strings.TrimSpace(res.Extra["projection"].(string)) != "" {
		t.Errorf("empty doc projection: status=%s proj=%q, want OK/empty", res.Status, res.Extra["projection"])
	}
}
