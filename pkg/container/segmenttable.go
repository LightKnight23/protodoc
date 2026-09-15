// SegmentTable: the fixed raw-struct array occupying container octets
// [262144,1048576) (contracts/container.abnf S5, data-model.md S2.4,
// FR-091). Like Header and CommitRingRecord, this region is zero-tag,
// fixed-layout (data-model.md S2.4 invariant list, item 5 in the shared
// "never PDL-TLV" note carried from frontmatter.go): every slot has a
// fixed offset and width, never a tag.
package container

import (
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// SegmentTable region sizing: the region spans [262144,1048576),
// immediately following Frontmatter and filling the remainder of the
// 1,048,576-octet fixed prefix exactly (data-model.md S2.4).
const (
	SegmentTableOffset     = FrontmatterOffset + FrontmatterRegionSize // 262144
	MaxSegments            = 16384                                     // MAX_SEGMENTS (TR-006, NFR-004)
	SegmentTableSlotSize   = 48                                        // 1+1+6+6+2+32 octets, container.abnf S5
	SegmentTableRegionSize = MaxSegments * SegmentTableSlotSize        // 786432

	// MaxFramesPerSegment bounds slot-frame-count (container.abnf S5
	// slot-frame-count comment); enforcement of this ceiling against a
	// decoded slot belongs to a later task, not this struct definition.
	MaxFramesPerSegment = 8192
)

// Segment type octets (container.abnf S5 slot-segment-type). 0 is the
// normative "unused slot" marker, never a real segment kind. 1-4 are the
// closed 4-value segment-type enum {CONTENT, RESOURCE, HISTORY, ATTEST}
// (TR-008): decodeSegmentTableSlot rejects every other octet value
// structurally, by range check alone, so the leading 1,048,576 octets
// never carry a construct a conforming reader would execute, interpret,
// or dereference as anything but one of these five closed values.
const (
	SegmentTypeUnused   = 0
	SegmentTypeContent  = 1
	SegmentTypeResource = 2
	SegmentTypeHistory  = 3
	SegmentTypeAttest   = 4
)

// ErrInvalidSegmentType is returned when a decoded slot-segment-type octet
// falls outside the closed enum {0 unused, 1 CONTENT, 2 RESOURCE,
// 3 HISTORY, 4 ATTEST} (TR-008).
var ErrInvalidSegmentType = errors.New("container: slot-segment-type outside closed enum {0..4}")

// SlotFlagCoverageHint is slot-flags bit0 (container.abnf S5): a coverage
// HINT only, never authoritative on its own (authoritative coverage is
// always the CoverageDescriptor inside the covering signature).
const SlotFlagCoverageHint = 1 << 0

// SegmentTableSlot is one 48-octet raw slot in the SegmentTable array
// (data-model.md S2.4). Ordinal is the slot's index in the table
// (0..MaxSegments-1); an all-zero slot (SegmentType == SegmentTypeUnused
// and every other field at its zero value) is the normative "unused"
// marker, not uninitialised padding.
type SegmentTableSlot struct {
	SegmentType byte             // slot-segment-type: 0=unused, 1=CONTENT, 2=RESOURCE, 3=HISTORY, 4=ATTEST
	Flags       byte             // slot-flags: bit0 = SlotFlagCoverageHint, bits1-7 reserved
	Offset      uint64           // slot-offset (u48): absolute file offset of the segment's first octet
	Length      uint64           // slot-length (u48): total octet length of the segment
	FrameCount  uint16           // slot-frame-count: number of frames in the segment
	Digest      pdlfmt.Digest256 // slot-digest: SHA-256 over the segment's complete octets
}

// ErrSegmentTableTruncated is returned when fewer than SegmentTableRegionSize octets are available.
var ErrSegmentTableTruncated = errors.New("container: segment table region truncated, need 786432 octets")

// ErrTooManySegments is returned when more than MaxSegments slots are supplied to EncodeSegmentTable.
var ErrTooManySegments = errors.New("container: more than MaxSegments (16384) slots supplied")

// encode writes s's canonical 48-octet encoding to dst, which must be
// exactly SegmentTableSlotSize octets (both len and cap, as supplied by
// EncodeSegmentTable's fixed-stride slicing): the in-place AppendUint48
// calls below rely on dst[2:2] and dst[8:8] each having at least 6 octets
// of spare capacity ahead of them, which a 48-octet window guarantees.
func (s *SegmentTableSlot) encode(dst []byte) error {
	if len(dst) != SegmentTableSlotSize {
		return fmt.Errorf("container: slot encode buffer is %d octets, want %d", len(dst), SegmentTableSlotSize)
	}
	dst[0] = s.SegmentType
	dst[1] = s.Flags
	if _, err := pdlfmt.AppendUint48(dst[2:2], s.Offset); err != nil {
		return fmt.Errorf("container: slot-offset: %w", err)
	}
	if _, err := pdlfmt.AppendUint48(dst[8:8], s.Length); err != nil {
		return fmt.Errorf("container: slot-length: %w", err)
	}
	dst[14] = byte(s.FrameCount >> 8)
	dst[15] = byte(s.FrameCount)
	copy(dst[16:48], s.Digest[:])
	return nil
}

// decodeSegmentTableSlot decodes one SegmentTableSlot from the leading
// SegmentTableSlotSize octets of src.
func decodeSegmentTableSlot(src []byte) (SegmentTableSlot, error) {
	var slot SegmentTableSlot
	if len(src) != SegmentTableSlotSize {
		return slot, fmt.Errorf("container: slot decode buffer is %d octets, want %d", len(src), SegmentTableSlotSize)
	}
	if src[0] > SegmentTypeAttest {
		return slot, fmt.Errorf("container: slot-segment-type %d: %w", src[0], ErrInvalidSegmentType)
	}
	slot.SegmentType = src[0]
	slot.Flags = src[1]
	offset, _, err := pdlfmt.DecodeUint48(src[2:8])
	if err != nil {
		return slot, fmt.Errorf("container: slot-offset: %w", err)
	}
	slot.Offset = offset
	length, _, err := pdlfmt.DecodeUint48(src[8:14])
	if err != nil {
		return slot, fmt.Errorf("container: slot-length: %w", err)
	}
	slot.Length = length
	slot.FrameCount = uint16(src[14])<<8 | uint16(src[15])
	digest, _, err := pdlfmt.DecodeDigest256(src[16:48])
	if err != nil {
		return slot, fmt.Errorf("container: slot-digest: %w", err)
	}
	slot.Digest = digest
	return slot, nil
}

// EncodeSegmentTable writes the canonical SegmentTableRegionSize-octet
// encoding of the segment table to dst, which must be at least
// SegmentTableRegionSize octets. slots holds the populated leading
// ordinals (0..len(slots)-1); every remaining ordinal up to MaxSegments
// is encoded as the normative all-zero "unused" slot. It returns
// dst[:SegmentTableRegionSize].
func EncodeSegmentTable(slots []SegmentTableSlot, dst []byte) ([]byte, error) {
	if len(slots) > MaxSegments {
		return nil, fmt.Errorf("%w: got %d", ErrTooManySegments, len(slots))
	}
	if cap(dst) < SegmentTableRegionSize {
		dst = make([]byte, SegmentTableRegionSize)
	} else {
		dst = dst[:SegmentTableRegionSize]
		for i := range dst {
			dst[i] = 0
		}
	}
	for i := range slots {
		if err := slots[i].encode(dst[i*SegmentTableSlotSize : (i+1)*SegmentTableSlotSize]); err != nil {
			return nil, fmt.Errorf("container: segment table slot %d: %w", i, err)
		}
	}
	return dst, nil
}

// DecodeSegmentTable decodes the full MaxSegments-slot array from the
// leading SegmentTableRegionSize octets of src. It reads no octet at or
// past SegmentTableRegionSize and performs no I/O of its own beyond
// reading src. Every slot's kind (SegmentType), length (Length) and
// digest (Digest) is readable from the returned slot alone (FR-091),
// with no separate index consulted.
func DecodeSegmentTable(src []byte) ([MaxSegments]SegmentTableSlot, error) {
	var table [MaxSegments]SegmentTableSlot
	if len(src) < SegmentTableRegionSize {
		return table, ErrSegmentTableTruncated
	}
	for i := 0; i < MaxSegments; i++ {
		slot, err := decodeSegmentTableSlot(src[i*SegmentTableSlotSize : (i+1)*SegmentTableSlotSize])
		if err != nil {
			return table, fmt.Errorf("container: segment table slot %d: %w", i, err)
		}
		table[i] = slot
	}
	return table, nil
}
