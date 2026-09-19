package cli

import (
	"os"
	"path/filepath"
	"testing"

	"Protodoc/pkg/container"
)

// writeDocWithSlotDigest writes a valid document with one CONTENT segment whose
// slot digest is set to digByte (so merge/diff can see distinct content
// identities), under the given history mode.
func writeDocWithSlotDigest(t *testing.T, mode container.HistoryMode, digByte byte) string {
	t.Helper()
	h := &container.Header{
		FormatMajor: 1, DocumentClass: 1, CapabilityWritten: 1, CapabilityRequired: 1,
		HistoryMode: mode, UnicodeVersionID: 1, PrefixLayoutID: 1,
	}
	var ring [container.CommitRingSlots]container.CommitRingRecord
	base := uint64(prefixSize)
	segLen := uint64(256)
	total := base + segLen
	for i := range ring {
		ring[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: total}
	}
	var dg [32]byte
	dg[0] = digByte
	slots := []container.SegmentTableSlot{
		{SegmentType: container.SegmentTypeContent, Offset: base, Length: segLen, FrameCount: 1, Digest: dg},
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
	img = append(img, make([]byte, segLen)...)
	path := filepath.Join(t.TempDir(), "d.pdl")
	if err := os.WriteFile(path, img, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

// TestTR_012_MergeVerbReadsRealFiles is T-0378's named integration test
// (TR-012/TR-003, DEFECT-2026-09-19 fix). The production merge backend
// (realMergeRun, wired by init) reads three real files: divergent edits to the
// same construct conflict; a history-mode mismatch is REFUSED (CON-025); an
// unchanged/clean set merges OK.
func TestTR_012_MergeVerbReadsRealFiles(t *testing.T) {
	// Base at ordinal-N with digest 0x10; side A changes it to 0x20, side B to
	// 0x30 -> divergent conflict.
	base := writeDocWithSlotDigest(t, container.HistoryComplete, 0x10)
	sideA := writeDocWithSlotDigest(t, container.HistoryComplete, 0x20)
	sideB := writeDocWithSlotDigest(t, container.HistoryComplete, 0x30)
	res := runMerge([]string{base, sideA, sideB}, nil)
	if res.Extra["result"] != "CONFLICT" {
		t.Errorf("divergent edits: result=%v, want CONFLICT", res.Extra["result"])
	}

	// History-mode mismatch between sides -> REFUSED CON-025.
	sideBNoHist := writeDocWithSlotDigest(t, container.HistoryNone, 0x20)
	res = runMerge([]string{base, sideA, sideBNoHist}, nil)
	if res.Status != "REFUSED" || res.Extra["condition"] != "CON-025" {
		t.Errorf("mode mismatch: status=%s condition=%v, want REFUSED/CON-025", res.Status, res.Extra["condition"])
	}

	// Both sides identical to base -> clean OK.
	res = runMerge([]string{base, base, base}, nil)
	if res.Status != "OK" {
		t.Errorf("no-op merge: status=%s, want OK", res.Status)
	}

	// Absent input -> USAGE, never REFUSED (DEFECT-2026-09-19b/T-0384: REFUSED
	// is reserved for a genuine policy refusal on an otherwise-fine input;
	// cli.md does not list exit code 6 among merge's used codes at all. And
	// per T-0390, an unreadable path is USAGE, not INVALID).
	res = runMerge([]string{base, sideA, filepath.Join(t.TempDir(), "absent.pdl")}, nil)
	if res.Status != "USAGE" {
		t.Errorf("absent input: status=%s, want USAGE", res.Status)
	}

	if r := runMerge([]string{base, sideA}, nil); r.Status != "USAGE" {
		t.Errorf("two-operand merge: status=%s, want USAGE", r.Status)
	}
}
