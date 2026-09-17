// Opaque-construct preservation on save (T-0039, FR-017): when a writer
// saves a document containing a construct it does not implement, it must
// reproduce that construct octet-for-octet in the output or refuse the
// save naming each construct it could not preserve. At the ledger layer
// the unit of an unrecognised construct is a sealed segment whose
// capability generation exceeds the writer's implemented generation: the
// writer cannot interpret such a segment's frames, so it must carry the
// segment's octets through unchanged (which append-only placement does by
// construction) rather than dropping, re-framing or approximating it.
//
// This file provides the save-path guard that turns that obligation into a
// checkable fact: given the segments a prior image declares and the writer's
// own capability generation, it identifies the opaque (unrecognised)
// segments and verifies each one's exact octets are present, unchanged, in
// the saved image -- refusing with a construct-naming error otherwise.
package ledger

import (
	"bytes"
	"errors"
	"fmt"

	"Protodoc/pkg/container"
)

// OpaqueSegment identifies one unrecognised, preserve-verbatim segment in a
// prior image: its storage ordinal and the octet range it occupies.
type OpaqueSegment struct {
	Ordinal uint64
	Offset  uint64
	Length  uint64
	// CapabilityGen is the segment's own seg-capability-gen, the generation
	// a writer would need to implement to interpret it; it exceeds the
	// saving writer's generation, which is why the segment is opaque.
	CapabilityGen uint16
}

// UnpreservedConstructError is FR-017's refusal: it names each opaque
// construct the save could not reproduce octet-for-octet. It is returned
// rather than silently dropping or approximating the construct.
type UnpreservedConstructError struct {
	// Ordinals are the storage ordinals of the opaque segments whose octets
	// were not carried through the save unchanged.
	Ordinals []uint64
}

func (e *UnpreservedConstructError) Error() string {
	return fmt.Sprintf("ledger: save would not preserve unimplemented construct(s) octet-for-octet: segment ordinal(s) %v (FR-017 refusal, not a silent drop)", e.Ordinals)
}

// ErrOpaqueSegmentOutOfBounds is returned when a declared opaque segment's
// [Offset, Offset+Length) range does not lie within the prior image: a
// malformed input caught before any comparison (CP-006).
var ErrOpaqueSegmentOutOfBounds = errors.New("ledger: opaque segment range lies outside the prior image")

// IdentifyOpaqueSegments returns the segments in slots that the writer does
// not implement: those whose seg-capability-gen exceeds writerCapabilityGen.
// seg-capability-gen is read from each segment's own 64-octet header in the
// prior image (container.abnf S6.1, offset 5..7 within the header), not from
// the SegmentTableSlot, since the slot carries no capability field. A slot
// whose segment octets are truncated in prior is skipped from the opaque set
// and reported via err so the caller can reject the malformed prior rather
// than mis-classify it.
func IdentifyOpaqueSegments(prior []byte, slots []container.SegmentTableSlot, writerCapabilityGen uint16) ([]OpaqueSegment, error) {
	var opaque []OpaqueSegment
	for i, s := range slots {
		if s.SegmentType == container.SegmentTypeUnused {
			continue
		}
		end := s.Offset + s.Length
		if s.Offset < PrefixLength || end < s.Offset || end > uint64(len(prior)) {
			return nil, fmt.Errorf("%w: ordinal %d range [%d,%d), image length %d", ErrOpaqueSegmentOutOfBounds, i, s.Offset, end, len(prior))
		}
		capGen := segmentCapabilityGen(prior[s.Offset:end])
		if capGen > writerCapabilityGen {
			opaque = append(opaque, OpaqueSegment{
				Ordinal:       uint64(i),
				Offset:        s.Offset,
				Length:        s.Length,
				CapabilityGen: capGen,
			})
		}
	}
	return opaque, nil
}

// segmentCapabilityGen reads seg-capability-gen (u16 at offset 5..7) from a
// segment's header octets. The segment slice is guaranteed >= 64 octets by
// IdentifyOpaqueSegments's bounds check before this is called; a shorter
// slice (which cannot be a real segment) reads as generation 0 so it is
// never misclassified as opaque.
func segmentCapabilityGen(seg []byte) uint16 {
	if len(seg) < 7 {
		return 0
	}
	return uint16(seg[5])<<8 | uint16(seg[6])
}

// VerifySavePreservesOpaque verifies that every opaque segment's exact
// octets from prior appear unchanged at the same offset in saved, and
// returns an UnpreservedConstructError naming every ordinal that does not.
// This is the FR-017 save-path assertion: a save that carries all opaque
// constructs through verbatim passes; one that drops, relocates or mutates
// any of them is refused with the construct named.
func VerifySavePreservesOpaque(prior, saved []byte, opaque []OpaqueSegment) error {
	var unpreserved []uint64
	for _, o := range opaque {
		priorOctets := prior[o.Offset : o.Offset+o.Length]
		end := o.Offset + o.Length
		if end > uint64(len(saved)) || !bytes.Equal(saved[o.Offset:end], priorOctets) {
			unpreserved = append(unpreserved, o.Ordinal)
		}
	}
	if len(unpreserved) > 0 {
		return &UnpreservedConstructError{Ordinals: unpreserved}
	}
	return nil
}
