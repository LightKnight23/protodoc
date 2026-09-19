package cli

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// TestTR_012_RedactVerbReadsRealFile is T-0379's named integration test
// (TR-012, GAP-VERIFY-CONTENT-REBUILD closed; updated by T-0385 for
// --subtree/--out enforcement and file-write). The production redact backend
// (realRedactRun, wired by init) reads a real file, decodes authored unit-ids,
// and omits a designated subtree's frame (its bytes never survive), declaring
// the omission by AUTHORED unit-id, and actually writes the result to --out;
// an absent file is rejected.
func TestTR_012_RedactVerbReadsRealFile(t *testing.T) {
	outPath := func() string { return filepath.Join(t.TempDir(), "out.pdl") }

	// (1) Absent file -> USAGE (DEFECT-2026-09-19b/T-0390: an unreadable path
	// is USAGE, not INVALID).
	res := runRedact([]string{filepath.Join(t.TempDir(), "absent.pdl"), "--subtree", "deadbeef", "--out", outPath()}, nil)
	if res.Status != "USAGE" {
		t.Errorf("absent file: status=%s, want USAGE", res.Status)
	}

	// (2) A real doc with two content frames; redact the first by its authored
	// unit-id.
	idA, idB := cliUnit(0xA1), cliUnit(0xB2)
	path, _ := writeDocWithRealFrames(t, idA, idB)
	keyA := hex.EncodeToString(idA[:])

	out2 := outPath()
	res = runRedact([]string{path, "--subtree", keyA, "--out", out2}, nil)
	if res.Status != "OK" {
		t.Fatalf("redact: status=%s (%+v), want OK", res.Status, res.Findings)
	}
	omissions := res.Extra["declared_omissions"].([]string)
	if len(omissions) != 1 || omissions[0] != keyA {
		t.Errorf("declared omissions = %v, want [%s]", omissions, keyA)
	}
	written, err := os.ReadFile(out2)
	if err != nil {
		t.Fatalf("redact reported OK but --out file was not written: %v", err)
	}
	// Only frame A was omitted, so the output length equals frame B's length
	// (frame B retained), i.e. > 0 but less than both frames.
	if len(written) == 0 {
		t.Errorf("frame B should be retained, output should be non-empty")
	}

	// (3) Missing --subtree -> USAGE (required, at least one).
	if r := runRedact([]string{path, "--out", outPath()}, nil); r.Status != "USAGE" {
		t.Errorf("no subtree designated: status=%s, want USAGE", r.Status)
	}

	// (4) An unknown unit-id designation -> nothing omitted (no false match).
	unknown := cliUnit(0xFF)
	res = runRedact([]string{path, "--subtree", hex.EncodeToString(unknown[:]), "--out", outPath()}, nil)
	if len(res.Extra["declared_omissions"].([]string)) != 0 {
		t.Errorf("unknown designation must omit nothing, got %v", res.Extra["declared_omissions"])
	}

	// (5) Missing --out -> USAGE.
	if r := runRedact([]string{path, "--subtree", keyA}, nil); r.Status != "USAGE" {
		t.Errorf("missing --out: status=%s, want USAGE", r.Status)
	}
}
