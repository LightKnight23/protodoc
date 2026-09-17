package ledger

import (
	"errors"
	"testing"

	"Protodoc/pkg/container"
)

// makeSegmentWithCapability builds a minimal synthetic sealed-segment
// payload of the given total length whose 64-octet header carries the
// given seg-capability-gen (container.abnf S6.1: seg-magic "PDS1" at [0,4),
// seg-type at [4], seg-capability-gen u16 at [5,7)). Only the fields
// IdentifyOpaqueSegments reads are set meaningfully; the rest is
// deterministic filler standing in for frames and directory, which this
// layer treats as opaque.
func makeSegmentWithCapability(segType byte, capGen uint16, length int) []byte {
	if length < 64 {
		length = 64
	}
	seg := make([]byte, length)
	copy(seg[0:4], []byte("PDS1"))
	seg[4] = segType
	seg[5] = byte(capGen >> 8)
	seg[6] = byte(capGen)
	for i := 64; i < length; i++ {
		seg[i] = byte(i*7 + 1)
	}
	return seg
}

// TestFR_017_OpaqueConstructPreservedOrRefused is T-0039's named test. It
// builds a document containing a synthetic segment whose capability
// generation exceeds the writer's (an unimplemented construct), then:
//
//  1. Preservation: an ordinary append-only save (place with an added
//     recognised segment) carries the opaque segment's octets through
//     byte-identical, and VerifySavePreservesOpaque accepts it.
//  2. Refusal: a fault-injected save whose image drops/mutates the opaque
//     segment's octets is refused with an UnpreservedConstructError naming
//     that segment's ordinal, never a silent drop.
func TestFR_017_OpaqueConstructPreservedOrRefused(t *testing.T) {
	const writerCapability = 1

	// Assemble a prior document: fixed prefix followed by two sealed
	// segments -- one recognised (capability 1) and one opaque
	// (capability 5, beyond the writer). The SegmentTable slots point at
	// their real offsets so IdentifyOpaqueSegments can read their headers.
	recognised := makeSegmentWithCapability(container.SegmentTypeContent, 1, 128)
	opaqueSeg := makeSegmentWithCapability(container.SegmentTypeResource, 5, 256)

	recOffset := uint64(PrefixLength)
	opaqueOffset := recOffset + uint64(len(recognised))

	slots := []container.SegmentTableSlot{
		{SegmentType: container.SegmentTypeContent, Offset: recOffset, Length: uint64(len(recognised)), FrameCount: 1},
		{SegmentType: container.SegmentTypeResource, Offset: opaqueOffset, Length: uint64(len(opaqueSeg)), FrameCount: 1},
	}

	var ring [container.CommitRingSlots]container.CommitRingRecord
	for i := range ring {
		ring[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: opaqueOffset + uint64(len(opaqueSeg)), SegmentCount: 2}
	}
	tail := append(append([]byte(nil), recognised...), opaqueSeg...)
	prior := corpusDocument(t, &ring, &container.Frontmatter{}, slots, tail)

	opaque, err := IdentifyOpaqueSegments(prior, slots, writerCapability)
	if err != nil {
		t.Fatalf("IdentifyOpaqueSegments: %v", err)
	}
	if len(opaque) != 1 || opaque[0].Ordinal != 1 {
		t.Fatalf("expected exactly one opaque segment at ordinal 1, got %+v", opaque)
	}

	// --- 1. Preservation via an ordinary append-only save ---
	added := makeSegmentWithCapability(container.SegmentTypeContent, 1, 96)
	res, err := place(prior, 2, EditDelta{NewSegments: []SegmentPayload{{Octets: added}}})
	if err != nil {
		t.Fatalf("place (edit save): %v", err)
	}
	if err := VerifySavePreservesOpaque(prior, res.Image, opaque); err != nil {
		t.Fatalf("append-only save did not preserve the opaque construct: %v", err)
	}
	// Explicit octet check for the opaque segment.
	got := res.Image[opaqueOffset : opaqueOffset+uint64(len(opaqueSeg))]
	for i := range got {
		if got[i] != opaqueSeg[i] {
			t.Fatalf("opaque segment octet %d changed after save: was 0x%02x, now 0x%02x", i, opaqueSeg[i], got[i])
		}
	}

	// --- 2. Refusal on a save that fails to carry the opaque construct ---
	// Fault-inject: build a "saved" image that truncates the opaque
	// segment's octets (a writer that dropped the construct it could not
	// interpret). The guard must refuse, naming the ordinal.
	dropped := append([]byte(nil), prior[:opaqueOffset]...) // everything up to but not including the opaque segment
	err = VerifySavePreservesOpaque(prior, dropped, opaque)
	if err == nil {
		t.Fatalf("a save that dropped the opaque construct was not refused")
	}
	var uce *UnpreservedConstructError
	if !errors.As(err, &uce) {
		t.Fatalf("refusal error is %T, want *UnpreservedConstructError", err)
	}
	if len(uce.Ordinals) != 1 || uce.Ordinals[0] != 1 {
		t.Fatalf("refusal named ordinals %v, want [1]", uce.Ordinals)
	}

	// A mutated (not dropped) opaque segment must also be refused.
	mutated := append([]byte(nil), res.Image...)
	mutated[opaqueOffset+70] ^= 0xFF // flip a byte inside the opaque segment
	if err := VerifySavePreservesOpaque(prior, mutated, opaque); err == nil {
		t.Fatalf("a save that mutated the opaque construct was not refused")
	}
}

// TestFR_017_RejectsMalformedOpaqueRange confirms IdentifyOpaqueSegments
// rejects a slot whose declared range lies outside the prior image before
// classifying it, rather than reading out of bounds (CP-006).
func TestFR_017_RejectsMalformedOpaqueRange(t *testing.T) {
	prior := make([]byte, PrefixLength+64)
	slots := []container.SegmentTableSlot{
		{SegmentType: container.SegmentTypeContent, Offset: PrefixLength, Length: 1 << 20}, // runs past end
	}
	if _, err := IdentifyOpaqueSegments(prior, slots, 1); !errors.Is(err, ErrOpaqueSegmentOutOfBounds) {
		t.Fatalf("out-of-bounds slot: got %v, want ErrOpaqueSegmentOutOfBounds", err)
	}
}
