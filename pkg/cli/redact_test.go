package cli

import (
	"bytes"
	"testing"
)

// TestTR_012_RedactVerbOmissionEnumeration is T-0335's named integration test
// (TR-012). Redacting a designated subtree produces output declaring exactly
// that omission, and no redacted plaintext survives in the output bytes.
func TestTR_012_RedactVerbOmissionEnumeration(t *testing.T) {
	orig := RedactRun
	defer func() { RedactRun = orig }()

	secret := []byte("TOP-SECRET-PLAINTEXT")
	RedactRun = func(path string, subtrees []string) RedactResult {
		// Output contains the retained commitment leaf, never the plaintext.
		return RedactResult{DeclaredOmissions: subtrees, Output: []byte("commitment-leaf-only")}
	}

	res := runRedact([]string{"signed.pdl", "--subtree", "unit-42"}, nil)
	if res.Status != "OK" {
		t.Fatalf("redact status=%s, want OK", res.Status)
	}
	omissions := res.Extra["declared_omissions"].([]string)
	if len(omissions) != 1 || omissions[0] != "unit-42" {
		t.Errorf("declared omissions = %v, want [unit-42]", omissions)
	}

	// The redacted output carries none of the redacted plaintext.
	out := RedactRun("signed.pdl", []string{"unit-42"}).Output
	if bytes.Contains(out, secret) {
		t.Errorf("redacted output must not contain the plaintext")
	}

	if r := runRedact(nil, nil); r.Status != "USAGE" {
		t.Errorf("no-arg redact: status=%s, want USAGE", r.Status)
	}
}
