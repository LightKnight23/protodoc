package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// realFrameWithPayload builds a real, decodable content-model frame carrying
// both the authored unit-id (tag=1, required for extract.LoadContentRecords)
// and a distinguishing payload (tag=2), so two frames sharing the same
// unit-id can still carry different content -- needed to simulate "the same
// construct, edited differently on each side" for merge tests.
func realFrameWithPayload(discriminant byte, id pdlfmt.UnitID, payload []byte) []byte {
	f := []byte{discriminant}
	f = pdlfmt.AppendField(f, pdlfmt.Field{Tag: 1, Value: id[:]})
	f = pdlfmt.AppendField(f, pdlfmt.Field{Tag: 2, Value: payload})
	return f
}

// writeDocWithFrames writes a valid document whose CONTENT segments are
// exactly the given pre-built frames, in order, under the given history mode.
func writeDocWithFrames(t *testing.T, mode container.HistoryMode, frames [][]byte) string {
	t.Helper()
	h := &container.Header{
		FormatMajor: 1, DocumentClass: 1, CapabilityWritten: 1, CapabilityRequired: 1,
		HistoryMode: mode, UnicodeVersionID: 1, PrefixLayoutID: 1,
	}
	var ring [container.CommitRingSlots]container.CommitRingRecord
	base := uint64(prefixSize)
	off := base
	slots := make([]container.SegmentTableSlot, len(frames))
	bodyTotal := uint64(0)
	for i, fr := range frames {
		slots[i] = container.SegmentTableSlot{
			SegmentType: container.SegmentTypeContent,
			Offset:      off, Length: uint64(len(fr)), FrameCount: 1,
		}
		off += uint64(len(fr))
		bodyTotal += uint64(len(fr))
	}
	total := base + bodyTotal
	for i := range ring {
		ring[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: total}
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
	for _, fr := range frames {
		img = append(img, fr...)
	}
	path := filepath.Join(t.TempDir(), "mergefixture.pdl")
	if err := os.WriteFile(path, img, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

// TestTR_012_MergeVerbReadsRealFiles is T-0378's named integration test
// (TR-012/TR-003, DEFECT-2026-09-19 fix; updated by T-0391/DEFECT-2026-09-19c
// for the real per-construct three-way merge and --out file-write). The
// production merge backend (realMergeRun, wired by init) reads three real
// files: divergent edits to the same construct conflict (real values named);
// a history-mode mismatch is REFUSED (CON-025); an unchanged/clean set merges
// OK and writes real output.
func TestTR_012_MergeVerbReadsRealFiles(t *testing.T) {
	id := cliUnit(0x10)
	outPath := func() string { return filepath.Join(t.TempDir(), "merged.pdl") }

	// Base at id with payload "base"; side A changes it to "left", side B to
	// "right" -> divergent conflict.
	base := writeDocWithFrames(t, container.HistoryComplete, [][]byte{realFrameWithPayload(0x01, id, []byte("base"))})
	sideA := writeDocWithFrames(t, container.HistoryComplete, [][]byte{realFrameWithPayload(0x01, id, []byte("left"))})
	sideB := writeDocWithFrames(t, container.HistoryComplete, [][]byte{realFrameWithPayload(0x01, id, []byte("right"))})
	res := runMerge([]string{base, sideA, sideB, "--out", outPath()}, nil)
	if res.Extra["result"] != "CONFLICT" {
		t.Errorf("divergent edits: result=%v, want CONFLICT", res.Extra["result"])
	}

	// History-mode mismatch between sides -> REFUSED CON-025.
	sideBNoHist := writeDocWithFrames(t, container.HistoryNone, [][]byte{realFrameWithPayload(0x01, id, []byte("left"))})
	res = runMerge([]string{base, sideA, sideBNoHist, "--out", outPath()}, nil)
	if res.Status != "REFUSED" || res.Extra["condition"] != "CON-025" {
		t.Errorf("mode mismatch: status=%s condition=%v, want REFUSED/CON-025", res.Status, res.Extra["condition"])
	}

	// Both sides identical to base -> clean OK, real output written.
	out3 := outPath()
	res = runMerge([]string{base, base, base, "--out", out3}, nil)
	if res.Status != "OK" {
		t.Errorf("no-op merge: status=%s, want OK", res.Status)
	}
	if written, err := os.ReadFile(out3); err != nil || len(written) == 0 {
		t.Errorf("no-op merge: --out not written correctly, content=%q err=%v", written, err)
	}

	// One side changed, the other unchanged -> clean OK, real output carries
	// the changed side's value.
	out4 := outPath()
	res = runMerge([]string{base, sideA, base, "--out", out4}, nil)
	if res.Status != "OK" {
		t.Errorf("one-sided change: status=%s, want OK", res.Status)
	}
	written4, err := os.ReadFile(out4)
	if err != nil {
		t.Fatalf("one-sided change: --out not written: %v", err)
	}
	// The merged canonical output should carry sideA's real "left" payload,
	// not base's "base" payload -- proving this is a genuine merge of real
	// content, not a stub or a copy of one input.
	if !bytes.Contains(written4, []byte("left")) {
		t.Errorf("one-sided change: merged output %q does not contain the real changed content \"left\"", written4)
	}

	// Absent input -> USAGE (T-0390: an unreadable path is USAGE, not
	// REFUSED/INVALID).
	res = runMerge([]string{base, sideA, filepath.Join(t.TempDir(), "absent.pdl"), "--out", outPath()}, nil)
	if res.Status != "USAGE" {
		t.Errorf("absent input: status=%s, want USAGE", res.Status)
	}

	if r := runMerge([]string{base, sideA}, nil); r.Status != "USAGE" {
		t.Errorf("two-operand merge: status=%s, want USAGE", r.Status)
	}
}
