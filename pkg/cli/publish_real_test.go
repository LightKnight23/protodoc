package cli

import (
	"path/filepath"
	"testing"
)

// TestTR_012_PublishVerbReadsRealFile is T-0380's named integration test
// (TR-012, DEFECT-2026-09-19 fix). The production publish backend
// (realPublishRun, wired by init) reads a real file: an absent/invalid file is
// rejected; a real document is re-emitted via the canon path with zero residue
// and custody preserved.
func TestTR_012_PublishVerbReadsRealFile(t *testing.T) {
	// (1) Absent file -> INVALID.
	res := runPublish([]string{filepath.Join(t.TempDir(), "absent.pdl")}, nil)
	if res.Status != "INVALID" {
		t.Errorf("absent file: status=%s, want INVALID", res.Status)
	}

	// (2) A real document with CONTENT segments -> OK, zero residue, custody OK.
	doc := writeDocWithContentSegments(t, 2, 256)
	res = runPublish([]string{doc}, nil)
	if res.Status != "OK" {
		t.Fatalf("publish: status=%s (%+v), want OK", res.Status, res.Findings)
	}
	if res.Extra["residue_octets"].(int) != 0 {
		t.Errorf("publish must have zero residue, got %d", res.Extra["residue_octets"])
	}
	if res.Extra["custody_preserved"] != true {
		t.Errorf("publish must preserve custody/fixity")
	}

	// (3) An empty valid document -> OK.
	empty := writeValidPrefix(t)
	if r := runPublish([]string{empty}, nil); r.Status != "OK" {
		t.Errorf("empty doc publish: status=%s, want OK", r.Status)
	}

	if r := runPublish(nil, nil); r.Status != "USAGE" {
		t.Errorf("no-arg publish: status=%s, want USAGE", r.Status)
	}
}
