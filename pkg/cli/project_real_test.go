package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTR_012_ProjectVerbReadsRealFile is T-0376's named integration test
// (TR-012/TR-004, DEFECT-2026-09-19 fix, updated by T-0383 for --to
// enforcement/file-write). The production project backend
// (realProjectStateFor, wired by init) reads real files: an absent/garbage file
// fails the CP-006 gate; a real document with CONTENT segments produces a
// non-empty, deterministic projection derived from the actual frame bytes and
// actually written to the required --to path.
func TestTR_012_ProjectVerbReadsRealFile(t *testing.T) {
	toPath := func() string { return filepath.Join(t.TempDir(), "out.txt") }

	// (1) Absent file -> INVALID (CP-006 validate-first).
	res := runProject([]string{filepath.Join(t.TempDir(), "absent.pdl"), "--to", toPath()}, nil)
	if res.Status != "INVALID" {
		t.Errorf("absent file: status=%s, want INVALID", res.Status)
	}

	// (2) A real document with CONTENT frames -> non-empty projection keyed by
	// the AUTHORED unit-id, actually written to --to.
	doc, ids := writeDocWithRealFrames(t, cliUnit(0xA1), cliUnit(0xB2))
	out1 := toPath()
	res = runProject([]string{doc, "--format=text", "--to", out1}, nil)
	if res.Status != "OK" {
		t.Fatalf("content doc: status=%s (%+v), want OK", res.Status, res.Findings)
	}
	if res.Extra["out"] != out1 {
		t.Errorf("result Extra[\"out\"]=%v, want %q", res.Extra["out"], out1)
	}
	projectionBytes, err := os.ReadFile(out1)
	if err != nil {
		t.Fatalf("project reported OK but --to file was not written: %v", err)
	}
	projection := string(projectionBytes)
	if strings.TrimSpace(projection) == "" {
		t.Errorf("projection of a document with content must be non-empty")
	}
	// The projection is keyed by the authored unit-id (hex of ids[0] appears).
	if !strings.Contains(projection, "a1") {
		t.Errorf("projection should include the authored unit-id (a1..), got %q", projection)
	}
	_ = ids

	// Deterministic: projecting twice gives identical output.
	out2 := toPath()
	res2 := runProject([]string{doc, "--format=text", "--to", out2}, nil)
	if res2.Status != "OK" {
		t.Fatalf("second run: status=%s, want OK", res2.Status)
	}
	projection2, err := os.ReadFile(out2)
	if err != nil {
		t.Fatalf("second run reported OK but --to file was not written: %v", err)
	}
	if string(projection2) != projection {
		t.Errorf("project output not deterministic across runs")
	}

	// html format also works over the real content.
	outH := toPath()
	resH := runProject([]string{doc, "--format=html", "--to", outH}, nil)
	htmlBytes, herr := os.ReadFile(outH)
	if resH.Status != "OK" || herr != nil || !strings.Contains(string(htmlBytes), "data-unit") {
		t.Errorf("html projection over real content failed: status=%s err=%v content=%q", resH.Status, herr, htmlBytes)
	}

	// (3) Empty doc -> empty projection file, still OK.
	empty := writeValidPrefix(t)
	outEmpty := toPath()
	res = runProject([]string{empty, "--format=text", "--to", outEmpty}, nil)
	emptyBytes, eerr := os.ReadFile(outEmpty)
	if res.Status != "OK" || eerr != nil || strings.TrimSpace(string(emptyBytes)) != "" {
		t.Errorf("empty doc projection: status=%s err=%v content=%q, want OK/empty file", res.Status, eerr, emptyBytes)
	}

	// (4) Missing --to -> USAGE, no output written anywhere unexpected.
	if r := runProject([]string{doc, "--format=text"}, nil); r.Status != "USAGE" {
		t.Errorf("missing --to: status=%s, want USAGE", r.Status)
	}
}
