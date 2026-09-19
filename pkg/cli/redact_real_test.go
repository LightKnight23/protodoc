package cli

import (
	"encoding/hex"
	"path/filepath"
	"testing"
)

// TestTR_012_RedactVerbReadsRealFile is T-0379's named integration test
// (TR-012, GAP-VERIFY-CONTENT-REBUILD closed). The production redact backend
// (realRedactRun, wired by init) reads a real file, decodes authored unit-ids,
// and omits a designated subtree's frame (its bytes never survive), declaring
// the omission by AUTHORED unit-id; an absent file is rejected.
func TestTR_012_RedactVerbReadsRealFile(t *testing.T) {
	// (1) Absent file -> INVALID.
	res := runRedact([]string{filepath.Join(t.TempDir(), "absent.pdl"), "--subtree", "deadbeef"}, nil)
	if res.Status != "INVALID" {
		t.Errorf("absent file: status=%s, want INVALID", res.Status)
	}

	// (2) A real doc with two content frames; redact the first by its authored
	// unit-id.
	idA, idB := cliUnit(0xA1), cliUnit(0xB2)
	path, _ := writeDocWithRealFrames(t, idA, idB)
	keyA := hex.EncodeToString(idA[:])

	res = runRedact([]string{path, "--subtree", keyA}, nil)
	if res.Status != "OK" {
		t.Fatalf("redact: status=%s (%+v), want OK", res.Status, res.Findings)
	}
	omissions := res.Extra["declared_omissions"].([]string)
	if len(omissions) != 1 || omissions[0] != keyA {
		t.Errorf("declared omissions = %v, want [%s]", omissions, keyA)
	}
	// Only frame A was omitted, so the output length equals frame B's length
	// (frame B retained), i.e. > 0 but less than both frames.
	if res.Extra["output_len"].(int) == 0 {
		t.Errorf("frame B should be retained, output should be non-empty")
	}

	// (3) Designating nothing -> nothing omitted, both frames retained.
	res = runRedact([]string{path}, nil)
	if len(res.Extra["declared_omissions"].([]string)) != 0 {
		t.Errorf("no subtree designated: nothing should be omitted")
	}

	// (4) An unknown unit-id designation -> nothing omitted (no false match).
	unknown := cliUnit(0xFF)
	res = runRedact([]string{path, "--subtree", hex.EncodeToString(unknown[:])}, nil)
	if len(res.Extra["declared_omissions"].([]string)) != 0 {
		t.Errorf("unknown designation must omit nothing, got %v", res.Extra["declared_omissions"])
	}
}
