package extract_test

import (
	"bytes"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/extract"
	"Protodoc/pkg/pdlfmt"
)

// buildContentFrame builds a minimal, well-formed content-model frame: a
// discriminant octet (tag 0) followed by a PDL-TLV record whose tag=1 field is
// the 16-octet authored unit-id. Ascending tag order is required by the TLV
// encoder, so tag 1 is the only field here.
func buildContentFrame(discriminant byte, id pdlfmt.UnitID) []byte {
	frame := []byte{discriminant}
	frame = pdlfmt.AppendField(frame, pdlfmt.Field{Tag: 1, Value: id[:]})
	return frame
}

// TestLoadContentRecordsDecodesAuthoredUnitID proves the whole-document decode
// path (GAP-VERIFY-CONTENT-REBUILD closer) reads each CONTENT frame's AUTHORED
// unit-id, not a slot-digest surrogate.
func TestLoadContentRecordsDecodesAuthoredUnitID(t *testing.T) {
	mkID := func(b byte) pdlfmt.UnitID {
		var id pdlfmt.UnitID
		id[0] = b
		id[15] = b ^ 0x5a
		return id
	}
	idA, idB := mkID(0xA1), mkID(0xB2)
	frameA := buildContentFrame(0x01, idA) // TEXT_BLOCK discriminant
	frameB := buildContentFrame(0x01, idB)

	// Assemble a real document: prefix declaring 2 CONTENT segments + bodies.
	h := &container.Header{FormatMajor: 1, DocumentClass: 1, CapabilityWritten: 1, CapabilityRequired: 1, HistoryMode: container.HistoryComplete, UnicodeVersionID: 1, PrefixLayoutID: 1}
	var ring [container.CommitRingSlots]container.CommitRingRecord
	prefixLen := container.SegmentTableOffset + container.SegmentTableRegionSize
	base := uint64(prefixLen)
	total := base + uint64(len(frameA)+len(frameB))
	for i := range ring {
		ring[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: total}
	}
	slots := []container.SegmentTableSlot{
		{SegmentType: container.SegmentTypeContent, Offset: base, Length: uint64(len(frameA)), FrameCount: 1},
		{SegmentType: container.SegmentTypeContent, Offset: base + uint64(len(frameA)), Length: uint64(len(frameB)), FrameCount: 1},
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
	if len(img) != prefixLen {
		t.Fatalf("prefix %d, want %d", len(img), prefixLen)
	}
	img = append(img, frameA...)
	img = append(img, frameB...)

	units, err := extract.LoadContentRecords(bytes.NewReader(img))
	if err != nil {
		t.Fatalf("LoadContentRecords: %v", err)
	}
	if len(units) != 2 {
		t.Fatalf("decoded %d content units, want 2", len(units))
	}
	if units[0].UnitID != idA || units[1].UnitID != idB {
		t.Errorf("authored unit-ids not decoded: got %x/%x, want %x/%x",
			units[0].UnitID[0], units[1].UnitID[0], idA[0], idB[0])
	}
	// The canonical frame octets are preserved exactly.
	if !bytes.Equal(units[0].Frame, frameA) {
		t.Errorf("frame octets not preserved for unit A")
	}

	// A malformed frame (discriminant only, no unit-id field) is a hard error.
	if _, err := extract.DecodeContentFrameUnitID([]byte{0x01}); err == nil {
		t.Errorf("a frame with no tag=1 unit-id field must error")
	}
}
