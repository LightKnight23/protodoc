package ledger

import (
	"bytes"
	"testing"

	"Protodoc/pkg/container"
)

// TestFR_018_OpaqueSegmentSurvivesTwoDivergentCopies is T-0040's named
// test. It establishes, at the ledger primitive level (ahead of M13's
// merge classifier), the storage-layer invariant FR-018's merge guarantee
// relies on: given two independently-edited copies of a document that each
// contain the same unrecognised construct, both copies retain that
// construct's octets unchanged after their respective independent edits.
// The merge algorithm itself is out of scope here (M13); this proves the
// bytes a merge would need are still present and identical in both inputs.
func TestFR_018_OpaqueSegmentSurvivesTwoDivergentCopies(t *testing.T) {
	const writerCapability = 1

	// A shared base document containing one opaque (capability-5) segment.
	opaqueSeg := makeSegmentWithCapability(container.SegmentTypeResource, 5, 320)
	opaqueOffset := uint64(PrefixLength)

	slots := []container.SegmentTableSlot{
		{SegmentType: container.SegmentTypeResource, Offset: opaqueOffset, Length: uint64(len(opaqueSeg)), FrameCount: 1},
	}
	var ring [container.CommitRingSlots]container.CommitRingRecord
	for i := range ring {
		ring[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: opaqueOffset + uint64(len(opaqueSeg)), SegmentCount: 1}
	}
	base := corpusDocument(t, &ring, &container.Frontmatter{}, slots, append([]byte(nil), opaqueSeg...))

	opaque, err := IdentifyOpaqueSegments(base, slots, writerCapability)
	if err != nil {
		t.Fatalf("IdentifyOpaqueSegments: %v", err)
	}
	if len(opaque) != 1 {
		t.Fatalf("expected one opaque segment in base, got %d", len(opaque))
	}

	// Fork into two copies (independent byte slices), then apply
	// UNRELATED independent edits to each via place(): copy A appends one
	// content segment, copy B appends two different segments. The base
	// already used ordinals 0..0, so both copies continue from ordinal 1.
	copyA := append([]byte(nil), base...)
	copyB := append([]byte(nil), base...)

	resA, err := place(copyA, 1, EditDelta{NewSegments: []SegmentPayload{
		{Octets: makeSegmentWithCapability(container.SegmentTypeContent, 1, 100)},
	}})
	if err != nil {
		t.Fatalf("place on copy A: %v", err)
	}
	resB, err := place(copyB, 1, EditDelta{NewSegments: []SegmentPayload{
		{Octets: makeSegmentWithCapability(container.SegmentTypeContent, 1, 150)},
		{Octets: makeSegmentWithCapability(container.SegmentTypeContent, 1, 175)},
	}})
	if err != nil {
		t.Fatalf("place on copy B: %v", err)
	}

	// The two edited copies must have diverged (different lengths from
	// different edits), so the retention result below is meaningful.
	if len(resA.Image) == len(resB.Image) {
		t.Fatalf("test setup error: the two copies did not diverge (both length %d)", len(resA.Image))
	}

	// The opaque segment's octets must be byte-identical in both edited
	// copies, and equal to the original.
	o := opaque[0]
	origOctets := base[o.Offset : o.Offset+o.Length]
	aOctets := resA.Image[o.Offset : o.Offset+o.Length]
	bOctets := resB.Image[o.Offset : o.Offset+o.Length]

	if !bytes.Equal(aOctets, origOctets) {
		t.Fatalf("opaque segment octets changed in copy A after its independent edit")
	}
	if !bytes.Equal(bOctets, origOctets) {
		t.Fatalf("opaque segment octets changed in copy B after its independent edit")
	}
	if !bytes.Equal(aOctets, bOctets) {
		t.Fatalf("opaque segment octets differ between the two divergent copies")
	}

	// And the save-path guard agrees for both copies independently.
	if err := VerifySavePreservesOpaque(base, resA.Image, opaque); err != nil {
		t.Fatalf("copy A failed opaque-preservation guard: %v", err)
	}
	if err := VerifySavePreservesOpaque(base, resB.Image, opaque); err != nil {
		t.Fatalf("copy B failed opaque-preservation guard: %v", err)
	}
}
