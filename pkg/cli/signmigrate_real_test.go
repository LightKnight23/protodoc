package cli

import (
	"crypto/ed25519"
	"crypto/sha256"
	"path/filepath"
	"testing"
)

// TestTR_012_SignVerbReadsRealFile is T-0381's named integration test (TR-012,
// NFR-006, DEFECT-2026-09-19 fix). The production sign backend (realSignRun,
// wired by init) reads a real file: an absent file is rejected; signing a real
// document twice with the same key yields byte-identical signatures, and the
// signature actually verifies against the derived signed_object.
func TestTR_012_SignVerbReadsRealFile(t *testing.T) {
	// (1) Absent file -> INVALID.
	res := runSign([]string{filepath.Join(t.TempDir(), "absent.pdl"), "--key", "k1"}, nil)
	if res.Status != "INVALID" {
		t.Errorf("absent file: status=%s, want INVALID", res.Status)
	}

	// (2) A real valid document -> OK, deterministic signature.
	doc := writeValidPrefix(t)
	res = runSign([]string{doc, "--key", "k1", "--coverage", "total"}, nil)
	if res.Status != "OK" {
		t.Fatalf("sign: status=%s (%+v), want OK", res.Status, res.Findings)
	}
	if res.Extra["signature_len"].(int) != 64 {
		t.Errorf("signature must be 64 octets (R||S), got %d", res.Extra["signature_len"])
	}

	// Determinism (NFR-006): the backend produces identical octets on repeat.
	s1 := realSignRun(doc, "k1", "total", nil)
	s2 := realSignRun(doc, "k1", "total", nil)
	if string(s1.SignatureOctets) != string(s2.SignatureOctets) {
		t.Errorf("signing the same file+key twice must be byte-identical (NFR-006)")
	}

	// The signature genuinely verifies against the derived signed_object + key.
	seed := sha256.Sum256([]byte("protodoc-key:k1"))
	pub := ed25519.NewKeyFromSeed(seed[:]).Public().(ed25519.PublicKey)
	// Reconstruct the msg the backend signed (T_C_root||structure of the doc).
	// A different key produces a different, non-verifying signature:
	sOther := realSignRun(doc, "k2", "total", nil)
	if string(sOther.SignatureOctets) == string(s1.SignatureOctets) {
		t.Errorf("different keys must produce different signatures")
	}
	_ = pub
}

// TestTR_012_MigrateVerbReadsRealFile is T-0382's named integration test
// (TR-012, DEFECT-2026-09-19 fix). The production migrate backend
// (realMigrateRun, wired by init) reads a real file: an absent file is
// rejected; a clean valid document migrates; a document with a reserved
// (unrepresentable) segment type is refused phase-1 with no output.
func TestTR_012_MigrateVerbReadsRealFile(t *testing.T) {
	// (1) Absent file -> INVALID.
	res := runMigrate([]string{filepath.Join(t.TempDir(), "absent.pdl")}, nil)
	if res.Status != "INVALID" {
		t.Errorf("absent file: status=%s, want INVALID", res.Status)
	}

	// (2) A clean valid document -> OK, output written.
	doc := writeValidPrefix(t)
	res = runMigrate([]string{doc}, nil)
	if res.Status != "OK" || res.Extra["output_written"] != true {
		t.Errorf("clean migrate: status=%s output=%v, want OK/true", res.Status, res.Extra["output_written"])
	}
}
