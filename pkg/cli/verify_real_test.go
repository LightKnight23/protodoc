package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestTR_012_VerifyVerbReadsRealFile is T-0374's named integration test
// (TR-012, DEFECT-2026-09-19 fix). The production verify backend
// (realVerifyRun, wired by init) reads real files: a garbage/absent file is
// unverified (CP-006 validate-first gate); a genuine valid prefix carrying no
// ATTEST segment yields zero signatures (an unsigned document), never a
// spurious OK for a file it never read.
func TestTR_012_VerifyVerbReadsRealFile(t *testing.T) {
	// The production backend must be wired, not the nil stub.
	// (A nil-returning stub would make an unsigned doc indistinguishable from
	// an unread one; the real backend reads the file to reach that conclusion.)

	// (1) Absent file -> unverified (fails the CP-006 validate precondition).
	res := runVerify([]string{filepath.Join(t.TempDir(), "absent.pdl")}, nil)
	if res.Status != "UNVERIFIED" {
		t.Errorf("absent file: status=%s, want UNVERIFIED", res.Status)
	}

	// (2) Garbage file -> unverified.
	garbage := filepath.Join(t.TempDir(), "garbage.pdl")
	if err := os.WriteFile(garbage, []byte("definitely not a protodoc document"), 0o600); err != nil {
		t.Fatal(err)
	}
	res = runVerify([]string{garbage}, nil)
	if res.Status != "UNVERIFIED" {
		t.Errorf("garbage file: status=%s, want UNVERIFIED", res.Status)
	}

	// (3) A genuine valid prefix with NO ATTEST segment -> zero signatures.
	// The verb reads the real file, decodes the (empty) segment table, finds no
	// ATTEST slot, and reports no signatures — resolving to OK (nothing to
	// disprove), but crucially derived from actually reading the file.
	valid := writeValidPrefix(t)
	res = runVerify([]string{valid}, nil)
	sigs, ok := res.Extra["signatures"].([]map[string]any)
	if !ok {
		t.Fatalf("verify envelope missing signatures payload: %+v", res.Extra)
	}
	if len(sigs) != 0 {
		t.Errorf("unsigned valid document should report zero signatures, got %d", len(sigs))
	}
	if res.Status != "OK" {
		t.Errorf("unsigned valid document: status=%s, want OK", res.Status)
	}
}
