package container

import (
	"bytes"
	"math/rand"
	"sync"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// canonicalEncoding is one iteration's complete fixed-prefix canonical
// octets: Header, CommitRing, Frontmatter and SegmentTable, each captured
// independently so a mismatch in any one region is reported by name.
type canonicalEncoding struct {
	header, ring, frontmatter, segmentTable []byte
}

// encodeCanonicalFixture encodes one fixed logical state (h, ring, fm,
// slots -- all read-only, shared across every call) through Header,
// CommitRing, Frontmatter and SegmentTable, in an order permuted by rnd.
// The permutation stands in for "process map/struct field iteration
// order": none of the four Encode paths actually iterates a map, so this
// is what lets the test vary the one thing under the test's own control
// (call order) while still exercising the same claim NFR-001 makes --
// that the four regions' canonical octets depend on the logical state
// alone, never on the order or interleaving of the calls producing them.
func encodeCanonicalFixture(t *testing.T, h *Header, ring *[CommitRingSlots]CommitRingRecord, fm *Frontmatter, slots []SegmentTableSlot, rnd *rand.Rand) canonicalEncoding {
	t.Helper()
	var out canonicalEncoding
	for _, step := range rnd.Perm(4) {
		switch step {
		case 0:
			out.header = append([]byte(nil), h.Encode(nil)...)
		case 1:
			out.ring = append([]byte(nil), EncodeCommitRing(ring, nil)...)
		case 2:
			enc, err := fm.Encode(nil)
			if err != nil {
				t.Fatalf("Frontmatter.Encode: %v", err)
			}
			out.frontmatter = append([]byte(nil), enc...)
		case 3:
			enc, err := EncodeSegmentTable(slots, nil)
			if err != nil {
				t.Fatalf("EncodeSegmentTable: %v", err)
			}
			out.segmentTable = append([]byte(nil), enc...)
		}
	}
	return out
}

// TestNFR_001_FixedPrefixCanonicalOctetsDeterministic is T-0030's named
// test. Implements: NFR-001.
//
// One fixed logical state is encoded through Header, CommitRing,
// Frontmatter and SegmentTable 100 times, concurrently across goroutines
// (so the Go runtime's own scheduling order -- itself a source of
// process-level nondeterminism NFR-001 must be immune to -- varies freely
// across iterations) and with each iteration's own per-region call order
// independently randomized (see encodeCanonicalFixture). Every one of the
// 100 results must be byte-identical to the first: the canonical octets
// for a fixed state have zero free parameters end to end across the fixed
// prefix, regardless of iteration or scheduling order.
func TestNFR_001_FixedPrefixCanonicalOctetsDeterministic(t *testing.T) {
	h := validHeader()

	var ring [CommitRingSlots]CommitRingRecord
	for i := range ring {
		ring[i] = CommitRingRecord{
			Sequence:             uint64(i + 1),
			LedgerLength:         4096 + uint64(i),
			SegmentCount:         3,
			RetentionPoint:       1,
			CompactionGeneration: 1,
		}
		ring[i].LedgerRoot[0] = byte(i + 1)
		ring[i].StateID[0] = byte(i + 1)
		ring[i].FrontmatterDigest[0] = byte(i + 1)
		ring[i].SegmentTableDigest[0] = byte(i + 1)
		ring[i].TCRoot[0] = byte(i + 1)
		ring[i].IndexRoute[0] = uint16(i + 1)
		ring[i].StructureDigest[0] = byte(i + 1)
	}

	fm := &Frontmatter{
		Title:           "NFR-001 Determinism Fixture",
		PageCount:       3,
		PageWidth:       8 * pdlfmt.UnitsPerInch,
		PageHeight:      11 * pdlfmt.UnitsPerInch,
		Language:        "en",
		ColourProfileID: ColourProfileSRGBD65,
	}

	slots := []SegmentTableSlot{
		{SegmentType: SegmentTypeContent, Flags: SlotFlagCoverageHint, Offset: 1048576, Length: 4096, FrameCount: 2},
		{SegmentType: SegmentTypeResource, Offset: 1052672, Length: 2048, FrameCount: 1},
		{SegmentType: SegmentTypeHistory, Offset: 1054720, Length: 512, FrameCount: 4},
	}
	for i := range slots {
		slots[i].Digest[0] = byte(i + 1)
	}

	const iterations = 100
	results := make([]canonicalEncoding, iterations)
	var wg sync.WaitGroup
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// A per-goroutine rand.Rand: math/rand's package-level
			// generator is not safe for concurrent use across goroutines
			// without its own locking, and this test wants each
			// iteration's permutation independent of the others anyway.
			rnd := rand.New(rand.NewSource(int64(i)))
			results[i] = encodeCanonicalFixture(t, h, &ring, fm, slots, rnd)
		}(i)
	}
	wg.Wait()

	want := results[0]
	for i := 1; i < iterations; i++ {
		got := results[i]
		if !bytes.Equal(got.header, want.header) {
			t.Fatalf("iteration %d: Header octets differ from iteration 0", i)
		}
		if !bytes.Equal(got.ring, want.ring) {
			t.Fatalf("iteration %d: CommitRing octets differ from iteration 0", i)
		}
		if !bytes.Equal(got.frontmatter, want.frontmatter) {
			t.Fatalf("iteration %d: Frontmatter octets differ from iteration 0", i)
		}
		if !bytes.Equal(got.segmentTable, want.segmentTable) {
			t.Fatalf("iteration %d: SegmentTable octets differ from iteration 0", i)
		}
	}
}
