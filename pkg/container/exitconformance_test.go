// Milestone M01 exit-gate conformance suite (T-0032): every other test in
// this package round-trips one region of the fixed prefix at a time. This
// file is the one place all four regions (Header, CommitRing, Frontmatter,
// SegmentTable) are stitched together at their real fixed offsets into one
// complete 1,048,576-octet buffer and round-tripped as a single unit,
// alongside the milestone's two named structural conformance vectors
// (PD-RING-001's tie-break, CON-010's frame-count boundary pair) re-run
// against values recovered from that assembled buffer rather than
// hand-built in isolation.
package container

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"Protodoc/pkg/ceilings"
	"Protodoc/pkg/pdlfmt"
)

// assembleFixedPrefix concatenates h, ring, fm and slots' canonical
// encodings in fixed prefix order (Header, CommitRing, Frontmatter,
// SegmentTable) and returns the result. It fails the test via t.Fatalf on
// any encode error rather than returning one, since every call site in
// this file supplies fixtures that are expected to encode successfully.
func assembleFixedPrefix(t *testing.T, h *Header, ring *[CommitRingSlots]CommitRingRecord, fm *Frontmatter, slots []SegmentTableSlot) []byte {
	t.Helper()
	prefix := make([]byte, 0, SegmentTableOffset+SegmentTableRegionSize)
	prefix = append(prefix, h.Encode(nil)...)
	prefix = append(prefix, EncodeCommitRing(ring, nil)...)

	fmEnc, err := fm.Encode(nil)
	if err != nil {
		t.Fatalf("assembleFixedPrefix: Frontmatter.Encode: %v", err)
	}
	prefix = append(prefix, fmEnc...)

	stEnc, err := EncodeSegmentTable(slots, nil)
	if err != nil {
		t.Fatalf("assembleFixedPrefix: EncodeSegmentTable: %v", err)
	}
	prefix = append(prefix, stEnc...)
	return prefix
}

// exitConformanceDoc is one corpus entry: a representative document's four
// fixed-prefix regions, named for its t.Run subtest.
type exitConformanceDoc struct {
	name        string
	header      *Header
	ring        [CommitRingSlots]CommitRingRecord
	frontmatter *Frontmatter
	slots       []SegmentTableSlot
}

// exitConformanceCorpus builds the milestone exit-gate's representative
// document set: a minimal (all-defaults) document, a typical document
// exercising every Frontmatter field and multiple segment kinds, and an
// at-the-MaxSegments-ceiling document (T-0016's own at-limit fixture,
// reused here as part of a full assembled prefix rather than an isolated
// SegmentTable region).
func exitConformanceCorpus() []exitConformanceDoc {
	var minimalRing [CommitRingSlots]CommitRingRecord
	for i := range minimalRing {
		minimalRing[i] = fixtureRingRecord(uint64(i+1), StateID{})
	}

	var typicalRing [CommitRingSlots]CommitRingRecord
	derived := StateID{0xAA}
	typicalRing[0] = fixtureRingRecord(1, StateID{})
	for i := 1; i < CommitRingSlots; i++ {
		typicalRing[i] = fixtureRingRecord(uint64(i+1), derived)
	}

	typicalFM := fixtureFrontmatter()
	typicalFM.PreviewKind = PreviewKindPLP1
	typicalFM.PreviewRaster = []byte{1, 2, 3, 4}
	typicalFM.PreviewDigest = ComputeFrontmatterPreviewDigest(typicalFM)
	typicalFM.ColourProfileID = ColourProfileSRGBD65
	typicalFM.CoverageSummary = []CoverageSummaryEntry{
		{
			Mode:                 CoverageModeSubset,
			CoveredRanges:        []SegmentRange{{Start: 0, End: 2}},
			UncoveredRanges:      []SegmentRange{{Start: 2, End: 3}},
			PrefixRegionsBitmask: CoverageBitHeader | CoverageBitFrontmatter,
		},
	}

	typicalSlots := []SegmentTableSlot{
		fixtureSlot(1), // SegmentTypeContent
		{SegmentType: SegmentTypeResource, Offset: 2097152, Length: 8192, FrameCount: 1, Digest: pdlfmt.Digest256{0x02}},
		{SegmentType: SegmentTypeHistory, Offset: 2105344, Length: 4096, FrameCount: 1, Digest: pdlfmt.Digest256{0x03}},
		{SegmentType: SegmentTypeAttest, Offset: 2109440, Length: 512, FrameCount: 1, Digest: pdlfmt.Digest256{0x04}},
	}

	var maxRing [CommitRingSlots]CommitRingRecord
	for i := range maxRing {
		maxRing[i] = fixtureRingRecord(uint64(i+1), StateID{})
	}
	maxSlots := make([]SegmentTableSlot, MaxSegments)
	for i := range maxSlots {
		maxSlots[i] = fixtureSlot(byte(i))
	}

	return []exitConformanceDoc{
		{
			name:        "minimal",
			header:      validHeader(),
			ring:        minimalRing,
			frontmatter: &Frontmatter{},
			slots:       nil,
		},
		{
			name:        "typical",
			header:      validHeader(),
			ring:        typicalRing,
			frontmatter: typicalFM,
			slots:       typicalSlots,
		},
		{
			name:        "at MaxSegments ceiling",
			header:      validHeader(),
			ring:        maxRing,
			frontmatter: &Frontmatter{},
			slots:       maxSlots,
		},
	}
}

// TestM01_FixedPrefixExitConformanceSuite is T-0032's named test: M01's
// milestone exit gate. It assembles the corpus above into the complete
// 1,048,576-octet fixed prefix, confirms byte-exact round-trip of the
// stitched-together whole (not just each region decoded in isolation, as
// every other test in this package already does), and re-confirms
// PD-RING-001's tie-break and CON-010's frame-count boundary pair hold
// once their inputs have been carried through a full prefix assembly and
// back out via decode, rather than constructed by hand in isolation.
//
// Implements: NFR-001, CON-010.
func TestM01_FixedPrefixExitConformanceSuite(t *testing.T) {
	prefixTotalSize := int(ceilings.MustMax("File prefix total size"))
	if prefixTotalSize != SegmentTableOffset+SegmentTableRegionSize {
		t.Fatalf("ceilings table 'File prefix total size' = %d, does not match SegmentTableOffset+SegmentTableRegionSize = %d",
			prefixTotalSize, SegmentTableOffset+SegmentTableRegionSize)
	}

	t.Run("corpus byte-exact round-trip", func(t *testing.T) {
		for _, doc := range exitConformanceCorpus() {
			doc := doc
			t.Run(doc.name, func(t *testing.T) {
				prefix := assembleFixedPrefix(t, doc.header, &doc.ring, doc.frontmatter, doc.slots)
				if len(prefix) != prefixTotalSize {
					t.Fatalf("assembled prefix length = %d, want %d", len(prefix), prefixTotalSize)
				}

				gotHeader, err := DecodeHeader(prefix[:HeaderSize])
				if err != nil {
					t.Fatalf("DecodeHeader: %v", err)
				}
				if *gotHeader != *doc.header {
					t.Errorf("Header round-trip mismatch:\n got  %+v\n want %+v", *gotHeader, *doc.header)
				}

				gotRing, err := DecodeCommitRing(prefix[HeaderSize : HeaderSize+CommitRingSize])
				if err != nil {
					t.Fatalf("DecodeCommitRing: %v", err)
				}
				if *gotRing != doc.ring {
					t.Errorf("CommitRing round-trip mismatch:\n got  %+v\n want %+v", *gotRing, doc.ring)
				}

				gotFM, err := DecodeFrontmatter(prefix[FrontmatterOffset : FrontmatterOffset+FrontmatterRegionSize])
				if err != nil {
					t.Fatalf("DecodeFrontmatter: %v", err)
				}
				if gotFM.PreviewKind != doc.frontmatter.PreviewKind ||
					!bytes.Equal(gotFM.PreviewRaster, doc.frontmatter.PreviewRaster) ||
					gotFM.PreviewDigest != doc.frontmatter.PreviewDigest ||
					gotFM.Title != doc.frontmatter.Title ||
					gotFM.PageCount != doc.frontmatter.PageCount ||
					gotFM.PageWidth != doc.frontmatter.PageWidth ||
					gotFM.PageHeight != doc.frontmatter.PageHeight ||
					gotFM.Language != doc.frontmatter.Language ||
					gotFM.ColourProfileID != doc.frontmatter.ColourProfileID ||
					!reflect.DeepEqual(gotFM.CoverageSummary, doc.frontmatter.CoverageSummary) {
					t.Errorf("Frontmatter round-trip mismatch:\n got  %+v\n want %+v", *gotFM, *doc.frontmatter)
				}

				gotTable, err := DecodeSegmentTable(prefix[SegmentTableOffset:])
				if err != nil {
					t.Fatalf("DecodeSegmentTable: %v", err)
				}
				for i := 0; i < MaxSegments; i++ {
					var want SegmentTableSlot
					if i < len(doc.slots) {
						want = doc.slots[i]
					}
					if gotTable[i] != want {
						t.Fatalf("SegmentTable slot %d mismatch: got %+v, want %+v", i, gotTable[i], want)
					}
				}

				// The whole-prefix round-trip's authoritative check: an
				// assembly built from the decoded values, region by region,
				// must reproduce the identical octets the corpus fixture
				// originally encoded — not merely equal-by-field Go values.
				reassembled := assembleFixedPrefix(t, gotHeader, gotRing, gotFM, gotTable[:])
				if !bytes.Equal(reassembled, prefix) {
					t.Fatal("re-assembled prefix octets differ from the original assembled encoding")
				}
			})
		}
	})

	// TestFR_117_PDRING001_EqualSequenceTie (ringwinner_test.go) already
	// proves the tie-break against a bare CommitRingSize buffer; this
	// re-runs the identical fixture embedded inside a full 1,048,576-octet
	// assembled prefix, so the milestone exit gate does not merely trust
	// that the two never diverge.
	t.Run("PD-RING-001 tie-break survives full-prefix assembly", func(t *testing.T) {
		var ring [CommitRingSlots]CommitRingRecord
		for i := range ring {
			ring[i] = fixtureRingRecord(uint64(i+1), StateID{})
		}
		const tiedSequence = 200
		ring[1].Sequence = tiedSequence
		ring[4].Sequence = tiedSequence

		prefix := assembleFixedPrefix(t, validHeader(), &ring, &Frontmatter{}, nil)
		ringRegion := prefix[HeaderSize : HeaderSize+CommitRingSize]
		fileLength := uint64(len(prefix)) + 1<<20

		_, index, err := SelectWinner(ringRegion, fileLength)
		if err == nil {
			t.Fatalf("SelectWinner: got winner at index %d inside assembled prefix, want PD-RING-001 tie rejection", index)
		}
		var tieErr *RingWinnerTieError
		if !errors.As(err, &tieErr) {
			t.Fatalf("SelectWinner: got %v, want *RingWinnerTieError", err)
		}
		wantIndices := []int{1, 4}
		if tieErr.Sequence != tiedSequence {
			t.Errorf("tie sequence = %d, want %d", tieErr.Sequence, tiedSequence)
		}
		if !reflect.DeepEqual(tieErr.SlotIndices, wantIndices) {
			t.Errorf("tie slot indices = %v, want %v", tieErr.SlotIndices, wantIndices)
		}
	})

	// TestCON_010_FrameCountBoundaryAndOverByOne (framebounds_test.go)
	// already proves the at-limit/over-limit pair against hand-built
	// values; this re-derives the same two segmentLength fixtures from a
	// SegmentTableSlot decoded back out of a full assembled prefix.
	t.Run("frame-count boundary at-limit and over-limit via assembled SegmentTable", func(t *testing.T) {
		var ring [CommitRingSlots]CommitRingRecord
		for i := range ring {
			ring[i] = fixtureRingRecord(uint64(i+1), StateID{})
		}

		const frameCount = uint16(MaxFramesPerSegment)
		required := uint64(frameCount)*frameDirEntrySize + frameDirTrailingDigestSize
		atLimitLength := required + segmentHeaderSize // exactly enough
		overLimitLength := atLimitLength - 1          // one octet short

		buildSlot := func(segLen uint64) SegmentTableSlot {
			return SegmentTableSlot{
				SegmentType: SegmentTypeContent,
				Offset:      SegmentTableOffset + SegmentTableRegionSize,
				Length:      segLen,
				FrameCount:  frameCount,
				Digest:      pdlfmt.Digest256{0xAB},
			}
		}

		atLimitPrefix := assembleFixedPrefix(t, validHeader(), &ring, &Frontmatter{}, []SegmentTableSlot{buildSlot(atLimitLength)})
		atLimitTable, err := DecodeSegmentTable(atLimitPrefix[SegmentTableOffset:])
		if err != nil {
			t.Fatalf("DecodeSegmentTable (at-limit): %v", err)
		}
		if err := CheckFrameDirectoryBounds(atLimitTable[0].FrameCount, atLimitTable[0].Length); err != nil {
			t.Errorf("at-limit fixture (frameCount=%d, segmentLength=%d) rejected after prefix round-trip: %v", frameCount, atLimitLength, err)
		}

		overLimitPrefix := assembleFixedPrefix(t, validHeader(), &ring, &Frontmatter{}, []SegmentTableSlot{buildSlot(overLimitLength)})
		overLimitTable, err := DecodeSegmentTable(overLimitPrefix[SegmentTableOffset:])
		if err != nil {
			t.Fatalf("DecodeSegmentTable (over-limit): %v", err)
		}
		err = CheckFrameDirectoryBounds(overLimitTable[0].FrameCount, overLimitTable[0].Length)
		if err == nil {
			t.Fatalf("over-limit fixture (frameCount=%d, segmentLength=%d, one octet short) accepted after prefix round-trip, want rejection", frameCount, overLimitLength)
		}
		if !errors.Is(err, ErrFrameDirectoryDoesNotFit) {
			t.Errorf("over-limit fixture: got error %v, want ErrFrameDirectoryDoesNotFit", err)
		}
	})
}
