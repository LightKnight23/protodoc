package cli

import (
	"os"
	"path/filepath"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// realFrame builds a minimal well-formed content-model frame: a discriminant
// octet (tag 0) followed by a tag=1 unit-id field.
func realFrame(discriminant byte, id pdlfmt.UnitID) []byte {
	f := []byte{discriminant}
	f = pdlfmt.AppendField(f, pdlfmt.Field{Tag: 1, Value: id[:]})
	return f
}

// cliUnit builds a distinct unit-id seeded by b.
func cliUnit(b byte) pdlfmt.UnitID {
	var id pdlfmt.UnitID
	id[0] = b
	id[15] = b ^ 0x5a
	return id
}

// writeDocWithRealFrames writes a valid document whose CONTENT segments carry
// genuine content-model frames (each decodable into an authored unit-id).
// Returns the path and the authored unit-ids in storage order.
func writeDocWithRealFrames(t *testing.T, ids ...pdlfmt.UnitID) (string, []pdlfmt.UnitID) {
	t.Helper()
	frames := make([][]byte, len(ids))
	for i, id := range ids {
		frames[i] = realFrame(0x01, id) // TEXT_BLOCK
	}
	h := &container.Header{
		FormatMajor: 1, DocumentClass: 1, CapabilityWritten: 1, CapabilityRequired: 1,
		HistoryMode: container.HistoryComplete, UnicodeVersionID: 1, PrefixLayoutID: 1,
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
	path := filepath.Join(t.TempDir(), "realframes.pdl")
	if err := os.WriteFile(path, img, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path, append([]pdlfmt.UnitID(nil), ids...)
}
