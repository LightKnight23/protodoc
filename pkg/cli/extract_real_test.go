package cli

import (
	"os"
	"path/filepath"
	"testing"

	"Protodoc/pkg/container"
)

// writeDocWithContentSegments writes a real document: a valid fixed prefix
// whose segment table declares n CONTENT segments, followed by their bodies.
// Returns the path and total size. It is a genuine file extract.Walk can read.
func writeDocWithContentSegments(t *testing.T, n int, segLen uint64) string {
	t.Helper()
	h := &container.Header{
		FormatMajor: 1, DocumentClass: 1, CapabilityWritten: 1, CapabilityRequired: 1,
		HistoryMode: container.HistoryComplete, UnicodeVersionID: 1, PrefixLayoutID: 1,
	}
	var ring [container.CommitRingSlots]container.CommitRingRecord
	base := uint64(prefixSize)
	total := base + uint64(n)*segLen
	for i := range ring {
		ring[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: total}
	}
	slots := make([]container.SegmentTableSlot, n)
	for i := 0; i < n; i++ {
		slots[i] = container.SegmentTableSlot{
			SegmentType: container.SegmentTypeContent,
			Offset:      base + uint64(i)*segLen,
			Length:      segLen,
			FrameCount:  1,
		}
	}
	img := make([]byte, 0, total)
	img = append(img, h.Encode(nil)...)
	img = append(img, container.EncodeCommitRing(&ring, nil)...)
	fmEnc, err := (&container.Frontmatter{}).Encode(nil)
	if err != nil {
		t.Fatalf("Frontmatter.Encode: %v", err)
	}
	img = append(img, fmEnc...)
	stEnc, err := container.EncodeSegmentTable(slots, nil)
	if err != nil {
		t.Fatalf("EncodeSegmentTable: %v", err)
	}
	img = append(img, stEnc...)
	if len(img) != prefixSize {
		t.Fatalf("prefix %d, want %d", len(img), prefixSize)
	}
	for i := 0; i < n; i++ {
		body := make([]byte, segLen)
		for j := range body {
			body[j] = byte('A' + i)
		}
		img = append(img, body...)
	}
	path := filepath.Join(t.TempDir(), "content.pdl")
	if err := os.WriteFile(path, img, 0o600); err != nil {
		t.Fatalf("write content fixture: %v", err)
	}
	return path
}

// TestTR_012_ExtractVerbReadsRealFile is T-0375's named integration test
// (TR-012, DEFECT-2026-09-19 fix). The production extract backend
// (realExtractRun, wired by init) reads real files: an absent file fails the
// CP-006 gate; a real document with CONTENT segments yields the right unit
// count and a genuine streaming fraction < 1 (the first unit is reachable
// before the whole file is read).
func TestTR_012_ExtractVerbReadsRealFile(t *testing.T) {
	// (1) Absent file -> USAGE (DEFECT-2026-09-19b/T-0390: an unreadable path
	// is USAGE, not INVALID).
	res := runExtract([]string{filepath.Join(t.TempDir(), "absent.pdl")}, nil)
	if res.Status != "USAGE" {
		t.Errorf("absent file: status=%s, want USAGE", res.Status)
	}

	// (2) A real document with 3 CONTENT segments -> 3 units, frac in (0,1].
	doc := writeDocWithContentSegments(t, 3, 4096)
	res = runExtract([]string{doc}, nil)
	if res.Status != "OK" {
		t.Fatalf("content doc: status=%s (%+v), want OK", res.Status, res.Findings)
	}
	units := res.Extra["units"].([]string)
	if len(units) != 3 {
		t.Errorf("expected 3 content units, got %d", len(units))
	}
	frac := res.Extra["first_unit_read_frac"].(float64)
	if frac <= 0 || frac > 1 {
		t.Errorf("streaming fraction = %v, want in (0,1]", frac)
	}
	// With the prefix (1 MiB) + first 4 KiB segment out of ~1 MiB+12 KiB total,
	// the first unit is reachable well before the whole file: frac < 1.
	if frac >= 1.0 {
		t.Errorf("first unit should be reachable before reading the whole file, frac=%v", frac)
	}

	// (3) A valid prefix with zero content -> zero units, frac 0.
	empty := writeValidPrefix(t)
	res = runExtract([]string{empty}, nil)
	if res.Status != "OK" || len(res.Extra["units"].([]string)) != 0 {
		t.Errorf("empty doc: status=%s units=%v, want OK/0", res.Status, res.Extra["units"])
	}
}
