package cli

import (
	"os"
	"path/filepath"
	"testing"

	"Protodoc/pkg/container"
)

// writeValidPrefix writes a genuine, decodable fixed-prefix document (header +
// commit ring + frontmatter + segment table, zero content segments) to a temp
// file and returns its path. It is a real file the production decode chain can
// read, not a stub.
func writeValidPrefix(t *testing.T) string {
	t.Helper()
	h := &container.Header{
		FormatMajor: 1, DocumentClass: 1, CapabilityWritten: 1, CapabilityRequired: 1,
		HistoryMode: container.HistoryComplete, UnicodeVersionID: 1, PrefixLayoutID: 1,
	}
	var ring [container.CommitRingSlots]container.CommitRingRecord
	for i := range ring {
		ring[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: prefixSize}
	}
	img := make([]byte, 0, prefixSize)
	img = append(img, h.Encode(nil)...)
	img = append(img, container.EncodeCommitRing(&ring, nil)...)
	fmEnc, err := (&container.Frontmatter{}).Encode(nil)
	if err != nil {
		t.Fatalf("Frontmatter.Encode: %v", err)
	}
	img = append(img, fmEnc...)
	stEnc, err := container.EncodeSegmentTable(nil, nil)
	if err != nil {
		t.Fatalf("EncodeSegmentTable: %v", err)
	}
	img = append(img, stEnc...)
	if len(img) != prefixSize {
		t.Fatalf("assembled prefix %d octets, want %d", len(img), prefixSize)
	}
	path := filepath.Join(t.TempDir(), "valid.pdl")
	if err := os.WriteFile(path, img, 0o600); err != nil {
		t.Fatalf("write valid fixture: %v", err)
	}
	return path
}

// TestTR_012_ValidateVerbRejectsGarbageFileAcceptsSample is T-0373's named
// integration test (TR-012, CP-006). It exercises the PRODUCTION validate
// backend (realValidateStepsFor, wired by init) against real files: an absent
// file, a garbage/truncated file, and a genuine valid prefix.
func TestTR_012_ValidateVerbRejectsGarbageFileAcceptsSample(t *testing.T) {
	// The production backend is wired by validate_real.go's init(); assert it.
	if _, checks := ValidateStepsFor("/nonexistent"); checks == 0 {
		t.Fatalf("production ValidateStepsFor must not be the no-op stub")
	}

	// (1) Absent file -> INVALID, not OK.
	res := runValidate([]string{filepath.Join(t.TempDir(), "does-not-exist.pdl")}, nil)
	if res.Status != "INVALID" {
		t.Errorf("absent file: status=%s, want INVALID", res.Status)
	}
	if len(res.Findings) == 0 {
		t.Errorf("absent file must produce a finding")
	}

	// (2) Garbage/truncated file (a few bytes) -> INVALID.
	garbage := filepath.Join(t.TempDir(), "garbage.pdl")
	if err := os.WriteFile(garbage, []byte("not a protodoc file at all"), 0o600); err != nil {
		t.Fatalf("write garbage: %v", err)
	}
	res = runValidate([]string{garbage}, nil)
	if res.Status != "INVALID" {
		t.Errorf("garbage file: status=%s, want INVALID", res.Status)
	}

	// (3) A genuine, decodable valid prefix -> OK.
	valid := writeValidPrefix(t)
	res = runValidate([]string{valid}, nil)
	if res.Status != "OK" {
		t.Errorf("valid sample: status=%s (findings %+v), want OK", res.Status, res.Findings)
	}

	// (4) The full binary path: a nonexistent file must NOT report OK (the
	// original defect). Verified here at the Dispatch level too.
	var stdout, stderr bytesBuffer
	code := Dispatch([]string{"validate", filepath.Join(t.TempDir(), "nope.pdl")}, &stdout, &stderr)
	if code == 0 {
		t.Errorf("Dispatch validate on absent file returned exit 0 (the defect); want non-zero")
	}
}

// bytesBuffer is a minimal io.Writer capturing output (avoids importing bytes
// twice across the package's test files).
type bytesBuffer struct{ b []byte }

func (w *bytesBuffer) Write(p []byte) (int, error) { w.b = append(w.b, p...); return len(p), nil }
