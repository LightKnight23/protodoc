package ledger

import (
	"testing"

	"Protodoc/pkg/container"
)

// maxSegmentsForTest mirrors the MAX_SEGMENTS ceiling (container.abnf S5)
// via the container package's own constant, so the ordinal-overflow test
// tracks the real ceiling rather than a restated literal.
const maxSegmentsForTest = uint64(container.MaxSegments)

// TestFR_058_StorageOrdinalMonotonicIndependentOfContent is T-0034's
// primary test. It establishes the two FR-058 / CQ-007-option-B
// properties place() must hold:
//
//  1. Storage ordinals are strictly increasing in allocation order and
//     never reused, even for byte-identical payloads placed at different
//     times: two segments with identical content receive different,
//     strictly increasing ordinals.
//  2. The ordinal a segment receives is a pure function of its allocation
//     position and derives from no name, digest or content bytes:
//     permuting the payloads' contents while holding their append order
//     fixed leaves the whole ordinal sequence unchanged.
func TestFR_058_StorageOrdinalMonotonicIndependentOfContent(t *testing.T) {
	// --- Property 1: byte-identical payloads get distinct, increasing
	// ordinals, and the sequence is strictly monotone across separate
	// place() calls that carry the running ordinal count forward. ---
	base := makePriorImage()

	// A single payload value reused for every append: identical content
	// must never collapse two ordinals together.
	identical := SegmentPayload{Octets: []byte("byte-identical-payload")}

	var issued []uint64
	image := base
	count := uint64(0)
	for edit := 0; edit < 8; edit++ {
		// Two identical segments per edit, so within one call two
		// byte-identical payloads must still get two different ordinals.
		res, err := place(image, count, EditDelta{NewSegments: []SegmentPayload{identical, identical}})
		if err != nil {
			t.Fatalf("edit %d: place: %v", edit, err)
		}
		for _, p := range res.Placements {
			issued = append(issued, p.Ordinal)
		}
		image = res.Image
		count += uint64(len(res.Placements))
	}

	for i := range issued {
		if uint64(i) != issued[i] {
			t.Fatalf("ordinal at allocation position %d is %d, want %d: ordinals are not the dense monotone allocation sequence", i, issued[i], uint64(i))
		}
		if i > 0 && issued[i] <= issued[i-1] {
			t.Fatalf("ordinal sequence not strictly increasing at position %d: %d then %d", i, issued[i-1], issued[i])
		}
	}

	// --- Property 2: content-independence. Place the SAME number of
	// segments in the SAME allocation order twice, once with one set of
	// contents and once with an unrelated, permuted set (different lengths
	// and different bytes), and assert the issued ordinal sequences are
	// identical. ---
	orderLen := 6
	runA := make([]SegmentPayload, orderLen)
	runB := make([]SegmentPayload, orderLen)
	for i := 0; i < orderLen; i++ {
		// runA: ascending lengths, one byte pattern.
		runA[i] = makeSegment(i, 10+i)
		// runB: descending lengths, an unrelated byte pattern, so both
		// content and size differ from runA at every position while the
		// append order (index 0..orderLen-1) is held fixed.
		runB[i] = makeSegment(1000-i, 10+(orderLen-1-i))
	}

	resA, err := place(makePriorImage(), 0, EditDelta{NewSegments: runA})
	if err != nil {
		t.Fatalf("place runA: %v", err)
	}
	resB, err := place(makePriorImage(), 0, EditDelta{NewSegments: runB})
	if err != nil {
		t.Fatalf("place runB: %v", err)
	}

	if len(resA.Placements) != orderLen || len(resB.Placements) != orderLen {
		t.Fatalf("placement counts %d/%d, want %d each", len(resA.Placements), len(resB.Placements), orderLen)
	}
	for i := 0; i < orderLen; i++ {
		if resA.Placements[i].Ordinal != resB.Placements[i].Ordinal {
			t.Fatalf("ordinal at position %d differs between two content sets: %d vs %d; ordinal is not content-independent", i, resA.Placements[i].Ordinal, resB.Placements[i].Ordinal)
		}
		if resA.Placements[i].Ordinal != uint64(i) {
			t.Fatalf("ordinal at position %d is %d, want %d", i, resA.Placements[i].Ordinal, uint64(i))
		}
	}

	// Sanity: the two runs really did carry different content and lengths,
	// so the equality above is a meaningful independence result, not two
	// identical inputs trivially agreeing.
	sawDifferentLength := false
	for i := 0; i < orderLen; i++ {
		if resA.Placements[i].Length != resB.Placements[i].Length {
			sawDifferentLength = true
			break
		}
	}
	if !sawDifferentLength {
		t.Fatalf("test setup error: runA and runB have identical lengths at every position, so content-independence is untested")
	}
}

// TestFR_058_OrdinalContinuesAcrossCallsAndOffsetTracksContent confirms
// the ordinal is carried forward correctly by priorSegmentCount across
// successive place() calls, and that Offset (unlike Ordinal) does depend
// on where the segment physically lands: the two are distinct concepts
// (ordinal = allocation position; offset = physical file location).
func TestFR_058_OrdinalContinuesAcrossCallsAndOffsetTracksContent(t *testing.T) {
	base := makePriorImage()

	res1, err := place(base, 0, EditDelta{NewSegments: []SegmentPayload{makeSegment(1, 50)}})
	if err != nil {
		t.Fatalf("first place: %v", err)
	}
	if res1.Placements[0].Ordinal != 0 {
		t.Fatalf("first segment ordinal %d, want 0", res1.Placements[0].Ordinal)
	}
	if res1.Placements[0].Offset != uint64(len(base)) {
		t.Fatalf("first segment offset %d, want %d (prior tail)", res1.Placements[0].Offset, uint64(len(base)))
	}
	if res1.Placements[0].Offset < PrefixLength {
		t.Fatalf("segment offset %d below PrefixLength %d", res1.Placements[0].Offset, uint64(PrefixLength))
	}

	// Next call carries priorSegmentCount = 1: the next ordinal must be 1.
	res2, err := place(res1.Image, 1, EditDelta{NewSegments: []SegmentPayload{makeSegment(2, 70)}})
	if err != nil {
		t.Fatalf("second place: %v", err)
	}
	if res2.Placements[0].Ordinal != 1 {
		t.Fatalf("second segment ordinal %d, want 1", res2.Placements[0].Ordinal)
	}
	if res2.Placements[0].Offset != uint64(len(res1.Image)) {
		t.Fatalf("second segment offset %d, want %d", res2.Placements[0].Offset, uint64(len(res1.Image)))
	}
}

// TestFR_058_RejectsOrdinalOverflowPastMaxSegments confirms place()
// refuses to issue a storage ordinal at or beyond MAX_SEGMENTS (16384):
// there is no SegmentTable slot to hold it (container.abnf S5). The check
// happens before allocation (CP-006), so no image is produced.
func TestFR_058_RejectsOrdinalOverflowPastMaxSegments(t *testing.T) {
	base := makePriorImage()

	// priorSegmentCount already at the ceiling: even one new segment
	// (which would want ordinal 16384) must be refused.
	_, err := place(base, maxSegmentsForTest, EditDelta{NewSegments: []SegmentPayload{makeSegment(1, 8)}})
	if err == nil {
		t.Fatalf("placing a segment at ordinal %d (== MAX_SEGMENTS) was not refused", maxSegmentsForTest)
	}

	// One below the ceiling: exactly one segment fits (ordinal 16383),
	// but two would push the second to 16384 and must be refused.
	if _, err := place(base, maxSegmentsForTest-1, EditDelta{NewSegments: []SegmentPayload{makeSegment(1, 8)}}); err != nil {
		t.Fatalf("placing the last legal ordinal (%d) was wrongly refused: %v", maxSegmentsForTest-1, err)
	}
	if _, err := place(base, maxSegmentsForTest-1, EditDelta{NewSegments: []SegmentPayload{makeSegment(1, 8), makeSegment(2, 8)}}); err == nil {
		t.Fatalf("placing two segments starting at ordinal %d (second would be %d == MAX_SEGMENTS) was not refused", maxSegmentsForTest-1, maxSegmentsForTest)
	}
}
