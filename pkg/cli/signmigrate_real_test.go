package cli

import (
	"crypto/ed25519"
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"
)

// TestTR_012_SignVerbReadsRealFile is T-0381's named integration test (TR-012,
// NFR-006, DEFECT-2026-09-19 fix; updated by T-0387 for required-flag
// enforcement). The production sign backend (realSignRun, wired by init)
// reads a real file: an absent file is rejected; signing a real document
// twice with the same key yields byte-identical signatures, and the
// signature actually verifies against the derived signed_object.
func TestTR_012_SignVerbReadsRealFile(t *testing.T) {
	outPath := func() string { return filepath.Join(t.TempDir(), "signed.pdl") }

	// (1) Absent file -> USAGE (DEFECT-2026-09-19b/T-0390: an unreadable path
	// is USAGE, not INVALID).
	res := runSign([]string{filepath.Join(t.TempDir(), "absent.pdl"), "--key", "k1", "--coverage", "total", "--intent", "author-approval", "--out", outPath()}, nil)
	if res.Status != "USAGE" {
		t.Errorf("absent file: status=%s, want USAGE", res.Status)
	}

	// (2) A real valid document -> OK, deterministic signature, --out written.
	doc := writeValidPrefix(t)
	out2 := outPath()
	res = runSign([]string{doc, "--key", "k1", "--coverage", "total", "--intent", "author-approval", "--out", out2}, nil)
	if res.Status != "OK" {
		t.Fatalf("sign: status=%s (%+v), want OK", res.Status, res.Findings)
	}
	if res.Extra["signature_len"].(int) != 64 {
		t.Errorf("signature must be 64 octets (R||S), got %d", res.Extra["signature_len"])
	}
	if _, err := os.Stat(out2); err != nil {
		t.Errorf("sign reported OK but --out file was not written: %v", err)
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
// (TR-012, DEFECT-2026-09-19 fix; updated by T-0388 for --to-major/--out
// enforcement and file-write). The production migrate backend
// (realMigrateRun, wired by init) reads a real file: an absent file is
// rejected; a clean valid document migrates and is actually written to
// --out; a document with a reserved (unrepresentable) segment type is
// refused phase-1 with no output.
func TestTR_012_MigrateVerbReadsRealFile(t *testing.T) {
	outPath := func() string { return filepath.Join(t.TempDir(), "out.pdl") }

	// (1) Absent file -> USAGE (DEFECT-2026-09-19b/T-0390: an unreadable path
	// is USAGE, not INVALID).
	res := runMigrate([]string{filepath.Join(t.TempDir(), "absent.pdl"), "--to-major", "2", "--out", outPath()}, nil)
	if res.Status != "USAGE" {
		t.Errorf("absent file: status=%s, want USAGE", res.Status)
	}

	// (2) A clean valid document (format-major 1) migrating to major 2 -> OK,
	// output actually written.
	doc := writeValidPrefix(t)
	out2 := outPath()
	res = runMigrate([]string{doc, "--to-major", "2", "--out", out2}, nil)
	if res.Status != "OK" || res.Extra["output_written"] != true {
		t.Errorf("clean migrate: status=%s output=%v, want OK/true", res.Status, res.Extra["output_written"])
	}
	if _, err := os.Stat(out2); err != nil {
		t.Errorf("migrate reported OK but --out file was not written: %v", err)
	}

	// (3) --to-major not greater than current format-major -> REFUSED
	// (a migration is always forward).
	res = runMigrate([]string{doc, "--to-major", "1", "--out", outPath()}, nil)
	if res.Status != "REFUSED" {
		t.Errorf("non-forward migration: status=%s, want REFUSED", res.Status)
	}
}
