package container

import (
	"errors"
	"testing"

	"Protodoc/pkg/ceilings"
)

// TestCON_010_MaxFramesPerSegmentMatchesGeneratedCeilingTable ties
// container.MaxFramesPerSegment (T-0016's existing struct-scoped
// constant, left unenforced there for this task per its own doc comment)
// to T-0017's generated ceiling table, so CheckFrameDirectoryBounds's
// ceiling really is "the generated ceiling constant" and not a second,
// independently drifting copy of the same number.
func TestCON_010_MaxFramesPerSegmentMatchesGeneratedCeilingTable(t *testing.T) {
	if got, want := uint64(MaxFramesPerSegment), ceilings.MustMax("MAX_FRAMES_PER_SEGMENT"); got != want {
		t.Errorf("container.MaxFramesPerSegment = %d, ceilings table MAX_FRAMES_PER_SEGMENT = %d", got, want)
	}
}

// TestCON_010_FrameCountBoundaryAndOverByOne is the CON-010 conformance
// vector pair plan.md's "Frame-directory offset arithmetic" threat entry
// calls out by name: frameCount held at MAX_FRAMES_PER_SEGMENT (the
// ceiling itself), with segmentLength first at the exact octet the frame
// directory needs (at-limit: accepted) and then one octet short of it
// (over-limit: rejected), with zero allocation performed in either case.
func TestCON_010_FrameCountBoundaryAndOverByOne(t *testing.T) {
	const frameCount = uint16(MaxFramesPerSegment) // 8192, the ceiling itself

	required := uint64(frameCount)*frameDirEntrySize + frameDirTrailingDigestSize // 393248
	atLimitLength := required + segmentHeaderSize                                 // 393312: exactly enough
	overLimitLength := atLimitLength - 1                                          // 393311: one octet short

	if err := CheckFrameDirectoryBounds(frameCount, atLimitLength); err != nil {
		t.Errorf("at-limit fixture (frameCount=%d, segmentLength=%d) rejected: %v", frameCount, atLimitLength, err)
	}

	err := CheckFrameDirectoryBounds(frameCount, overLimitLength)
	if err == nil {
		t.Fatalf("over-limit fixture (frameCount=%d, segmentLength=%d, one octet short) was accepted, want rejection", frameCount, overLimitLength)
	}
	if !errors.Is(err, ErrFrameDirectoryDoesNotFit) {
		t.Errorf("over-limit fixture: got error %v, want ErrFrameDirectoryDoesNotFit", err)
	}
}

// TestCON_010_FrameCountAboveCeilingRejectedIndependentlyOfLength confirms
// a frame count over MAX_FRAMES_PER_SEGMENT is rejected before any
// segmentLength arithmetic, even given an implausibly huge segment length
// that would otherwise satisfy the directory-fits inequality.
func TestCON_010_FrameCountAboveCeilingRejectedIndependentlyOfLength(t *testing.T) {
	overCeiling := uint16(MaxFramesPerSegment + 1) // 8193
	hugeLength := uint64(1) << 40

	err := CheckFrameDirectoryBounds(overCeiling, hugeLength)
	if err == nil {
		t.Fatalf("frameCount=%d (over MAX_FRAMES_PER_SEGMENT=%d) was accepted", overCeiling, MaxFramesPerSegment)
	}
	if !errors.Is(err, ErrFrameCountExceedsCeiling) {
		t.Errorf("got error %v, want ErrFrameCountExceedsCeiling", err)
	}
}

// TestCON_010_ShortSegmentLengthRejectedWithoutUnderflow confirms a
// segmentLength shorter than the fixed 64-octet segment header is
// rejected directly, rather than computing segmentLength-64 and wrapping
// to a huge unsigned value that would then falsely satisfy the bound.
func TestCON_010_ShortSegmentLengthRejectedWithoutUnderflow(t *testing.T) {
	err := CheckFrameDirectoryBounds(0, segmentHeaderSize-1)
	if err == nil {
		t.Fatal("segmentLength shorter than the segment header was accepted")
	}
	if !errors.Is(err, ErrFrameDirectoryDoesNotFit) {
		t.Errorf("got error %v, want ErrFrameDirectoryDoesNotFit", err)
	}
}

// TestCON_010_AcceptPathPerformsZeroAllocations confirms the DoD's "zero
// large allocation performed" property on the accepting path directly:
// the check itself never allocates anything, let alone anything
// proportional to frameCount, when the directory fits. (The rejecting
// path legitimately allocates the small, fixed-size error value fmt.Errorf
// builds; what CON-010 rules out is an allocation proportional to the
// attacker-controlled frameCount before rejection, which neither path
// ever performs — the check runs on the caller's own two integers, never
// on a buffer sized from them.)
func TestCON_010_AcceptPathPerformsZeroAllocations(t *testing.T) {
	const frameCount = uint16(MaxFramesPerSegment)
	required := uint64(frameCount)*frameDirEntrySize + frameDirTrailingDigestSize
	atLimitLength := required + segmentHeaderSize

	allocs := testing.AllocsPerRun(1000, func() {
		_ = CheckFrameDirectoryBounds(frameCount, atLimitLength)
	})
	if allocs != 0 {
		t.Errorf("CheckFrameDirectoryBounds (accept path) allocated %.0f times per run, want 0", allocs)
	}
}
