// Bounds-safe SegmentTableSlot frame-directory arithmetic (contracts/
// container.abnf S5 slot-frame-count NORMATIVE comment; S6.3 frame-
// directory NORMATIVE comment; data-model.md S2.4 item 3; plan.md
// "Frame-directory offset arithmetic" threat entry, HC-008/FR-106): a
// segment's frame directory (frame-dir-entry array + trailing
// segment-self-digest) must fit inside the segment's declared length,
// checked before that offset is computed or anything proportional to
// frame_count is allocated.
package container

import (
	"errors"
	"fmt"
)

// Sizes fixed by contracts/container.abnf, not present as their own rows
// in data-model.md's Section 5 ceiling table (SegmentTable slot size
// there names the unrelated 48-octet SegmentTableSlot struct, which is
// only coincidentally the same width as frame-dir-entry): named here
// directly from the ABNF rather than left as inline magic numbers.
const (
	// frameDirEntrySize is frame-dir-entry's width (container.abnf S6.3:
	// dir-discriminant(1) + dir-reserved(3) + dir-offset(6) +
	// dir-length(6) + dir-digest(32) = 48 octets).
	frameDirEntrySize = 48

	// frameDirTrailingDigestSize is the frame directory's one trailing
	// segment-self-digest (container.abnf S6.3): a digest256, 32 octets.
	frameDirTrailingDigestSize = 32

	// segmentHeaderSize is segment-header's fixed width (container.abnf
	// S6.1: seg-magic(4) + seg-type(1) + seg-capability-gen(2) +
	// seg-header-reserved(1) + seg-frame-count(4) + seg-payload-length(6)
	// + seg-tail-reserved(46) = 64 octets).
	segmentHeaderSize = 64
)

// ErrFrameCountExceedsCeiling is returned when a declared frame count
// exceeds MAX_FRAMES_PER_SEGMENT, before any arithmetic on segmentLength
// is attempted.
var ErrFrameCountExceedsCeiling = errors.New("container: slot-frame-count exceeds MAX_FRAMES_PER_SEGMENT")

// ErrFrameDirectoryDoesNotFit is returned when the frame directory that
// frameCount implies (frameCount*frameDirEntrySize +
// frameDirTrailingDigestSize octets) does not fit within the segment
// body (segmentLength - segmentHeaderSize octets).
var ErrFrameDirectoryDoesNotFit = errors.New("container: frame directory does not fit within declared segment length")

// CheckFrameDirectoryBounds implements the bounds-safe check
// container.abnf S5 requires of every live SegmentTableSlot before its
// segment's frame-directory offset is computed or any read or allocation
// proportional to frameCount is performed:
//
//	frameCount*48 + 32 <= segmentLength - 64
//
// It performs no allocation and reads no field but the two given: it is
// meant to run directly against a decoded SegmentTableSlot's FrameCount
// and Length, before touching the segment's own octets at all.
//
// frameCount is checked against MAX_FRAMES_PER_SEGMENT first,
// independent of segmentLength, since a frame count over that ceiling is
// invalid regardless of how large the segment claims to be. The
// remaining arithmetic is bounds-safe against unsigned underflow: a
// segmentLength shorter than segmentHeaderSize (which cannot hold even
// an empty frame directory) is rejected before segmentLength-64 is ever
// computed, rather than computed and silently wrapping to a huge value.
func CheckFrameDirectoryBounds(frameCount uint16, segmentLength uint64) error {
	if uint64(frameCount) > MaxFramesPerSegment {
		return fmt.Errorf("%w: got %d, ceiling %d", ErrFrameCountExceedsCeiling, frameCount, uint64(MaxFramesPerSegment))
	}
	if segmentLength < segmentHeaderSize {
		return fmt.Errorf("%w: segment length %d octets is shorter than the %d-octet segment header alone", ErrFrameDirectoryDoesNotFit, segmentLength, uint64(segmentHeaderSize))
	}
	available := segmentLength - segmentHeaderSize
	required := uint64(frameCount)*frameDirEntrySize + frameDirTrailingDigestSize
	if required > available {
		return fmt.Errorf("%w: frame directory needs %d octets, segment leaves %d after its %d-octet header (segment length %d)", ErrFrameDirectoryDoesNotFit, required, available, uint64(segmentHeaderSize), segmentLength)
	}
	return nil
}
