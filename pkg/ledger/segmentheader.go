// Ledger segment-header decode over untrusted bytes (T-0047 fuzz target's
// decode path; container.abnf S6.1 segment-header). This is the ledger's
// own trust boundary for the fixed 64-octet segment header that prefaces
// every sealed segment: it verifies before it trusts (CP-006), returns an
// ordinary structured error on any malformed input, and never panics,
// allocates from an unvalidated length, or reads outside the input buffer.
// The continuous-fuzzing target FuzzFR_056_SegmentFrameDecode feeds it
// arbitrary bytes (CP-012).
package ledger

import (
	"bytes"
	"errors"
	"fmt"

	"Protodoc/pkg/container"
)

// SegmentHeaderSize is the fixed segment-header width (container.abnf S6.1:
// seg-magic(4)+seg-type(1)+seg-capability-gen(2)+seg-header-reserved(1)+
// seg-frame-count(4)+seg-payload-length(6)+seg-tail-reserved(46) = 64).
const SegmentHeaderSize = 64

// SegMagic is the segment self-identification magic "PDS1" (container.abnf
// S6.1 seg-magic).
var SegMagic = [4]byte{'P', 'D', 'S', '1'}

// SegmentHeader is a decoded ledger segment header.
type SegmentHeader struct {
	Type          byte
	CapabilityGen uint16
	FrameCount    uint32
	PayloadLength uint64 // u48
}

var (
	// ErrSegmentHeaderTruncated is returned when fewer than
	// SegmentHeaderSize octets are available.
	ErrSegmentHeaderTruncated = errors.New("ledger: segment header truncated, need 64 octets")
	// ErrSegmentBadMagic is returned when seg-magic is not "PDS1".
	ErrSegmentBadMagic = errors.New("ledger: segment header magic mismatch (want PDS1)")
	// ErrSegmentBadType is returned when seg-type is outside the closed
	// enum {1 CONTENT, 2 RESOURCE, 3 HISTORY, 4 ATTEST}; 0 (unused) is not
	// a legal type for a real, sealed segment's own header.
	ErrSegmentBadType = errors.New("ledger: segment header type outside closed enum {1..4}")
	// ErrSegmentHeaderReservedNonzero is returned when seg-header-reserved
	// or seg-tail-reserved carries a nonzero octet (MBZ, no silent skip at
	// the segment-header level per container.abnf S6.1).
	ErrSegmentHeaderReservedNonzero = errors.New("ledger: segment header reserved octet is nonzero (MBZ)")
	// ErrSegmentFrameCountExceedsCeiling is returned when seg-frame-count
	// exceeds MAX_FRAMES_PER_SEGMENT.
	ErrSegmentFrameCountExceedsCeiling = errors.New("ledger: segment frame count exceeds MAX_FRAMES_PER_SEGMENT")
)

// DecodeSegmentHeader decodes and validates a ledger segment header from
// the leading SegmentHeaderSize octets of src. It reads no octet at or past
// SegmentHeaderSize, performs no allocation proportional to any decoded
// length, and returns a structured error (never a panic) for every
// malformed input: this is the CP-006 verify-before-trust boundary the
// CP-012 fuzz target exercises.
//
// Validation order: length, magic, type enum, reserved-MBZ octets,
// frame-count ceiling. seg-payload-length is decoded but not
// cross-validated against a segment length here (that is the
// SegmentTableSlot's bounds check, container.abnf S5); this function
// concerns only the header's own well-formedness.
func DecodeSegmentHeader(src []byte) (SegmentHeader, error) {
	var h SegmentHeader
	if len(src) < SegmentHeaderSize {
		return h, ErrSegmentHeaderTruncated
	}
	src = src[:SegmentHeaderSize]

	if !bytes.Equal(src[0:4], SegMagic[:]) {
		return h, fmt.Errorf("%w: got %x", ErrSegmentBadMagic, src[0:4])
	}

	segType := src[4]
	if segType < container.SegmentTypeContent || segType > container.SegmentTypeAttest {
		return h, fmt.Errorf("%w: got %d", ErrSegmentBadType, segType)
	}

	capGen := uint16(src[5])<<8 | uint16(src[6])

	// seg-header-reserved (offset 7) MBZ.
	if src[7] != 0 {
		return h, fmt.Errorf("%w: seg-header-reserved", ErrSegmentHeaderReservedNonzero)
	}

	frameCount := uint32(src[8])<<24 | uint32(src[9])<<16 | uint32(src[10])<<8 | uint32(src[11])
	if uint64(frameCount) > container.MaxFramesPerSegment {
		return h, fmt.Errorf("%w: got %d, ceiling %d", ErrSegmentFrameCountExceedsCeiling, frameCount, uint64(container.MaxFramesPerSegment))
	}

	// seg-payload-length is a u48 at offset 12..18 (big-endian).
	var payloadLen uint64
	for i := 12; i < 18; i++ {
		payloadLen = payloadLen<<8 | uint64(src[i])
	}

	// seg-tail-reserved (offset 18..64) MBZ.
	for i := 18; i < SegmentHeaderSize; i++ {
		if src[i] != 0 {
			return h, fmt.Errorf("%w: seg-tail-reserved offset %d", ErrSegmentHeaderReservedNonzero, i)
		}
	}

	h = SegmentHeader{
		Type:          segType,
		CapabilityGen: capGen,
		FrameCount:    frameCount,
		PayloadLength: payloadLen,
	}
	return h, nil
}

// EncodeSegmentHeader writes h's canonical 64-octet segment header to a new
// slice. It is the inverse of DecodeSegmentHeader for well-formed headers,
// used to seed the fuzz corpus and by callers assembling a segment.
func EncodeSegmentHeader(h SegmentHeader) []byte {
	out := make([]byte, SegmentHeaderSize)
	copy(out[0:4], SegMagic[:])
	out[4] = h.Type
	out[5] = byte(h.CapabilityGen >> 8)
	out[6] = byte(h.CapabilityGen)
	// out[7] header-reserved stays 0.
	out[8] = byte(h.FrameCount >> 24)
	out[9] = byte(h.FrameCount >> 16)
	out[10] = byte(h.FrameCount >> 8)
	out[11] = byte(h.FrameCount)
	pl := h.PayloadLength
	for i := 17; i >= 12; i-- {
		out[i] = byte(pl)
		pl >>= 8
	}
	// out[18:64] tail-reserved stays 0.
	return out
}
