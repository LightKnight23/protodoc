// Corpus definitions for CorpusVersion: the plan.md Section 10 first-class
// task fixtures (PD-RING-001's tie vector, the CON-010 frame_count
// boundary at-limit/over-limit pair), built from real container-package
// encodings rather than placeholder bytes, so a comparison run against
// this corpus actually exercises the fixed-prefix wire shapes NFR-028
// gates.
package diffconform

import (
	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// ringTieVector returns CommitRingSize raw octets encoding a 7-slot
// CommitRing in which slots 1 and 4 share the highest sequence number
// (200), the same tie construction TestFR_117_PDRING001_EqualSequenceTie
// (ringwinner_test.go) and TestM01_FixedPrefixExitConformanceSuite
// (exitconformance_test.go) already exercise in isolation: this corpus
// case re-poses it as a diffable byte vector two independent
// implementations' decoders must both reject the same way (PD-RING-001).
func ringTieVector() []byte {
	var ring [container.CommitRingSlots]container.CommitRingRecord
	for i := range ring {
		ring[i] = container.CommitRingRecord{Sequence: uint64(i + 1)}
	}
	const tiedSequence = 200
	ring[1].Sequence = tiedSequence
	ring[4].Sequence = tiedSequence
	return container.EncodeCommitRing(&ring, nil)
}

// frameCountBoundaryVector returns the SegmentTableRegionSize raw octets
// of a SegmentTable whose single live slot's declared frame count and
// segment length sit exactly at (overLimit=false) or one octet short of
// (overLimit=true) the CON-010 boundary
// frameCount*48+32 <= segmentLength-64, matching
// TestCON_010_FrameCountBoundaryAndOverByOne (framebounds_test.go) and
// TestM01_FixedPrefixExitConformanceSuite's re-derivation of the same
// pair from a full assembled prefix.
func frameCountBoundaryVector(overLimit bool) []byte {
	const frameCount = uint16(container.MaxFramesPerSegment)
	const frameDirEntrySize = 48
	const frameDirTrailingDigestSize = 32
	const segmentHeaderSize = 64

	required := uint64(frameCount)*frameDirEntrySize + frameDirTrailingDigestSize
	segLen := required + segmentHeaderSize // exactly enough (at-limit)
	if overLimit {
		segLen-- // one octet short
	}

	slot := container.SegmentTableSlot{
		SegmentType: container.SegmentTypeContent,
		Offset:      container.SegmentTableOffset + container.SegmentTableRegionSize,
		Length:      segLen,
		FrameCount:  frameCount,
		Digest:      pdlfmt.Digest256{0xAB},
	}

	enc, err := container.EncodeSegmentTable([]container.SegmentTableSlot{slot}, nil)
	if err != nil {
		// Corpus() is only ever called with statically-valid slot counts
		// (1, far under MaxSegments); a real error here is a programming
		// error in this file, not a runtime condition callers must
		// handle.
		panic("diffconform: frameCountBoundaryVector: EncodeSegmentTable: " + err.Error())
	}
	return enc
}

// Corpus returns this package's built-in conformance corpus at
// CorpusVersion: the two named first-class fixtures plan.md Section 10
// requires as their own dedicated cases (the PD-RING-001 tie vector and
// the CON-010 frame_count boundary vector, the latter as its required
// at-limit/over-limit pair). Adding, removing or changing any case here
// requires bumping CorpusVersion.
func Corpus() []CorpusCase {
	return []CorpusCase{
		{Name: "PD-RING-001 equal-sequence tie vector", Input: ringTieVector()},
		{Name: "CON-010 frame_count boundary: at-limit", Input: frameCountBoundaryVector(false)},
		{Name: "CON-010 frame_count boundary: over-limit", Input: frameCountBoundaryVector(true)},
	}
}
