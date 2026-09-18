package extract_test

import (
	"bytes"
	"io"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/extract"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_036_ExtractWalksRootSequenceOrder is T-0276's named integration test
// (FR-036, FR-101). It builds a document whose CONTENT segments are laid out in
// one storage order but whose authored ROOT_SEQUENCE reads them in a different
// order — the case of a structural edit that inserted a unit mid-document — and
// asserts the extraction walk emits text in reading order, not storage order.
func TestFR_036_ExtractWalksRootSequenceOrder(t *testing.T) {
	h := &container.Header{FormatMajor: 1, DocumentClass: 1, CapabilityWritten: 1, CapabilityRequired: 1, HistoryMode: container.HistoryComplete, UnicodeVersionID: 1, PrefixLayoutID: 1}
	var ring [container.CommitRingSlots]container.CommitRingRecord
	for i := range ring {
		ring[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: PrefixLenConst}
	}

	base := uint64(PrefixLenConst)
	segLen := uint64(64)
	// Three CONTENT segments in storage order 0,1,2 carrying 'A','B','C'.
	slots := []container.SegmentTableSlot{
		{SegmentType: container.SegmentTypeContent, Offset: base + 0*segLen, Length: segLen, FrameCount: 1},
		{SegmentType: container.SegmentTypeContent, Offset: base + 1*segLen, Length: segLen, FrameCount: 1},
		{SegmentType: container.SegmentTypeContent, Offset: base + 2*segLen, Length: segLen, FrameCount: 1},
	}

	img := make([]byte, 0, PrefixLenConst+3*int(segLen))
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
	if len(img) != PrefixLenConst {
		t.Fatalf("assembled prefix %d, want %d", len(img), PrefixLenConst)
	}
	markers := []byte{'A', 'B', 'C'}
	for _, m := range markers {
		img = append(img, bytes.Repeat([]byte{m}, int(segLen))...)
	}
	r := bytes.NewReader(img)

	// Map each CONTENT segment (by storage ordinal) to its unit-id.
	unit := func(seed byte) pdlfmt.UnitID {
		var id pdlfmt.UnitID
		id[0] = seed
		return id
	}
	uA, uB, uC := unit(0xA), unit(0xB), unit(0xC)
	unitOf := func(seg extract.ContentSegment) pdlfmt.UnitID {
		switch seg.Ordinal {
		case 0:
			return uA
		case 1:
			return uB
		default:
			return uC
		}
	}
	// Authored reading order C, A, B (a mid-document insertion put C first).
	order := []pdlfmt.UnitID{uC, uA, uB}

	var got []byte
	err = extract.WalkInReadingOrder(r, unitOf, order, func(seg extract.ContentSegment, sr io.Reader) error {
		buf := make([]byte, 1)
		if _, e := io.ReadFull(sr, buf); e != nil {
			return e
		}
		got = append(got, buf[0])
		return nil
	})
	if err != nil {
		t.Fatalf("WalkInReadingOrder: %v", err)
	}

	if string(got) != "CAB" {
		t.Errorf("emitted %q, want %q (reading order, not storage order 'ABC')", got, "CAB")
	}
}
