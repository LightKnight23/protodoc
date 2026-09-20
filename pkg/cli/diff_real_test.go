package cli

import (
	"os"
	"path/filepath"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// writeDocWithDecodableContent writes a valid document whose CONTENT segments
// are genuine decodable content-model frames (each carrying its own authored
// unit-id at tag=1, via realFrameWithPayload), one per given id/payload pair
// -- diff_real.go's real backend decodes CONTENT via
// extract.LoadContentRecords, which requires this shape (T-0395: the previous
// raw-filler fixture here could not exercise the real per-construct diff at
// all).
func writeDocWithDecodableContent(t *testing.T, ids []pdlfmt.UnitID, payloads [][]byte) string {
	t.Helper()
	frames := make([][]byte, len(ids))
	for i, id := range ids {
		frames[i] = realFrameWithPayload(0x01, id, payloads[i])
	}
	h := &container.Header{
		FormatMajor: 1, DocumentClass: 1, CapabilityWritten: 1, CapabilityRequired: 1,
		HistoryMode: container.HistoryComplete, UnicodeVersionID: 1, PrefixLayoutID: 1,
	}
	var ring [container.CommitRingSlots]container.CommitRingRecord
	base := uint64(prefixSize)
	off := base
	slots := make([]container.SegmentTableSlot, len(frames))
	for i, fr := range frames {
		slots[i] = container.SegmentTableSlot{SegmentType: container.SegmentTypeContent, Offset: off, Length: uint64(len(fr)), FrameCount: 1}
		off += uint64(len(fr))
	}
	total := off
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
	path := filepath.Join(t.TempDir(), "difffixture.pdl")
	if err := os.WriteFile(path, img, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

// TestTR_012_DiffVerbReadsRealFiles is T-0377's named integration test
// (TR-012/TR-002, DEFECT-2026-09-19 fix; fixture replaced by T-0395 to use
// real decodable content -- see diff_real.go's own doc comment for the defect
// this closes). The production diff backend (realDiffRun, wired by init)
// reads two real files: identical documents report zero changed constructs; a
// document whose real per-construct CONTENT differs (same segment count)
// reports that construct as changed; an absent input yields a diff-error
// sentinel, never a silent no-difference.
func TestTR_012_DiffVerbReadsRealFiles(t *testing.T) {
	idX := cliUnit(0x40)
	idY := cliUnit(0x41)

	// (1) Two identical real documents -> zero changed constructs.
	a := writeDocWithDecodableContent(t, []pdlfmt.UnitID{idX, idY}, [][]byte{[]byte("alpha"), []byte("beta")})
	b := writeDocWithDecodableContent(t, []pdlfmt.UnitID{idX, idY}, [][]byte{[]byte("alpha"), []byte("beta")})
	res := runDiff([]string{a, b}, nil)
	if res.Status != "OK" {
		t.Fatalf("identical docs: status=%s, want OK", res.Status)
	}
	if res.Extra["change_count"].(int) != 0 {
		t.Errorf("identical documents should report 0 changes, got %d", res.Extra["change_count"])
	}

	// (2) Same construct count, real CONTENT differs on one construct -> a
	// real change is reported (the exact bug T-0395 fixes: the previous
	// slot-digest comparison never populated a real digest, so this case was
	// wrongly reported identical).
	c := writeDocWithDecodableContent(t, []pdlfmt.UnitID{idX, idY}, [][]byte{[]byte("ALPHA-CHANGED"), []byte("beta")})
	res = runDiff([]string{a, c}, nil)
	if res.Extra["change_count"].(int) == 0 {
		t.Errorf("documents differing in real content must report changes, got 0")
	}
	if res.Extra["identical"] != false {
		t.Errorf("documents differing in real content: identical=%v, want false", res.Extra["identical"])
	}

	// (3) Absent input -> USAGE (DEFECT-2026-09-19b/T-0390: an unreadable path
	// is USAGE, not INVALID; the verb still refuses rather than reporting a
	// spurious no-difference OK).
	res = runDiff([]string{a, filepath.Join(t.TempDir(), "absent.pdl")}, nil)
	if res.Status != "USAGE" {
		t.Errorf("absent input must yield USAGE, got %s", res.Status)
	}

	// Missing operand -> USAGE.
	if r := runDiff([]string{a}, nil); r.Status != "USAGE" {
		t.Errorf("one-operand diff: status=%s, want USAGE", r.Status)
	}
}
