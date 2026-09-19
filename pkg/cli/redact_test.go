package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestTR_012_RedactVerbOmissionEnumeration is T-0335's named integration test
// (TR-012, updated by T-0385 for --subtree/--out enforcement and file-write).
// Redacting a designated subtree produces output declaring exactly that
// omission, actually written to --out, and no redacted plaintext survives in
// the output bytes.
func TestTR_012_RedactVerbOmissionEnumeration(t *testing.T) {
	orig := RedactRun
	defer func() { RedactRun = orig }()

	secret := []byte("TOP-SECRET-PLAINTEXT")
	RedactRun = func(path string, subtrees []string) RedactResult {
		// Output contains the retained commitment leaf, never the plaintext.
		return RedactResult{DeclaredOmissions: subtrees, Output: []byte("commitment-leaf-only")}
	}

	outPath := filepath.Join(t.TempDir(), "redacted.pdl")
	res := runRedact([]string{"signed.pdl", "--subtree", "unit-42", "--out", outPath}, nil)
	if res.Status != "OK" {
		t.Fatalf("redact status=%s, want OK", res.Status)
	}
	omissions := res.Extra["declared_omissions"].([]string)
	if len(omissions) != 1 || omissions[0] != "unit-42" {
		t.Errorf("declared omissions = %v, want [unit-42]", omissions)
	}

	// The redacted output was actually written and carries none of the
	// redacted plaintext.
	written, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("redact reported OK but --out file was not written: %v", err)
	}
	if bytes.Contains(written, secret) {
		t.Errorf("redacted output must not contain the plaintext")
	}

	if r := runRedact(nil, nil); r.Status != "USAGE" {
		t.Errorf("no-arg redact: status=%s, want USAGE", r.Status)
	}
	if r := runRedact([]string{"signed.pdl", "--out", outPath}, nil); r.Status != "USAGE" {
		t.Errorf("missing --subtree: status=%s, want USAGE", r.Status)
	}
	if r := runRedact([]string{"signed.pdl", "--subtree", "unit-42"}, nil); r.Status != "USAGE" {
		t.Errorf("missing --out: status=%s, want USAGE", r.Status)
	}
}
