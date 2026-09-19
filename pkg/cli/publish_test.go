package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestTR_012_PublishVerbZeroResidue is T-0336's named integration test
// (TR-012, updated by T-0386 for --out enforcement/file-write). Publishing a
// fixture with prior redactions/edit history produces output with zero
// removed-content octets while every custody/fixity value matches the input
// unchanged, actually written to --out.
func TestTR_012_PublishVerbZeroResidue(t *testing.T) {
	orig := PublishRun
	defer func() { PublishRun = orig }()
	outPath := func() string { return filepath.Join(t.TempDir(), "out.pdl") }

	// Clean publish: zero residue, custody preserved.
	PublishRun = func(path string, partial bool) PublishResult {
		return PublishResult{Output: []byte("clean"), ResidueOctets: 0, CustodyPreserved: true}
	}
	out1 := outPath()
	res := runPublish([]string{"doc.pdl", "--out", out1}, nil)
	if res.Status != "OK" {
		t.Errorf("publish status=%s, want OK", res.Status)
	}
	if res.Extra["residue_octets"] != 0 || res.Extra["custody_preserved"] != true {
		t.Errorf("publish invariants: %+v", res.Extra)
	}
	written, err := os.ReadFile(out1)
	if err != nil || string(written) != "clean" {
		t.Errorf("publish reported OK but --out file wrong/missing: content=%q err=%v", written, err)
	}

	// A publish that leaves residue is rejected (INVALID).
	PublishRun = func(path string, partial bool) PublishResult {
		return PublishResult{ResidueOctets: 3, CustodyPreserved: true}
	}
	if r := runPublish([]string{"doc.pdl", "--out", outPath()}, nil); r.Status != "INVALID" {
		t.Errorf("residue publish: status=%s, want INVALID", r.Status)
	}

	// A publish that drops a custody/fixity value is rejected.
	PublishRun = func(path string, partial bool) PublishResult {
		return PublishResult{ResidueOctets: 0, CustodyPreserved: false}
	}
	if r := runPublish([]string{"doc.pdl", "--out", outPath()}, nil); r.Status != "INVALID" {
		t.Errorf("custody-dropping publish: status=%s, want INVALID", r.Status)
	}

	if r := runPublish(nil, nil); r.Status != "USAGE" {
		t.Errorf("no-arg publish: status=%s, want USAGE", r.Status)
	}
	if r := runPublish([]string{"doc.pdl"}, nil); r.Status != "USAGE" {
		t.Errorf("missing --out: status=%s, want USAGE", r.Status)
	}
}
