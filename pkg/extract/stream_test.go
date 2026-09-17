package extract_test

import (
	"bytes"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/extract"
)

// buildContentDoc assembles a document with n CONTENT segments each of the
// given length, payload filled so segment k's octets are byte value k+1.
// Returns a tracking reader and the per-segment octet ranges.
func buildContentDoc(t *testing.T, n int, segLen int) (*trackingReaderAt, [][2]int64) {
	t.Helper()
	h := &container.Header{FormatMajor: 1, DocumentClass: 1, CapabilityWritten: 1, CapabilityRequired: 1, HistoryMode: container.HistoryComplete, UnicodeVersionID: 1, PrefixLayoutID: 1}
	var ring [container.CommitRingSlots]container.CommitRingRecord
	for i := range ring {
		ring[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: PrefixLenConst}
	}
	base := uint64(PrefixLenConst)
	slots := make([]container.SegmentTableSlot, n)
	ranges := make([][2]int64, n)
	for k := 0; k < n; k++ {
		off := base + uint64(k*segLen)
		slots[k] = container.SegmentTableSlot{SegmentType: container.SegmentTypeContent, Offset: off, Length: uint64(segLen), FrameCount: 1}
		ranges[k] = [2]int64{int64(off), int64(off) + int64(segLen)}
	}
	img := make([]byte, 0, PrefixLenConst+n*segLen)
	img = append(img, h.Encode(nil)...)
	img = append(img, container.EncodeCommitRing(&ring, nil)...)
	fmEnc, _ := (&container.Frontmatter{}).Encode(nil)
	img = append(img, fmEnc...)
	stEnc, _ := container.EncodeSegmentTable(slots, nil)
	img = append(img, stEnc...)
	for k := 0; k < n; k++ {
		img = append(img, bytes.Repeat([]byte{byte(k + 1)}, segLen)...)
	}
	return &trackingReaderAt{data: img}, ranges
}

// TestFR_047_EmitsEarlierUnitTextBeforeReadingLaterUnitOctets is T-0094's
// named integration test. For a 3-unit fixture, it pulls unit 1's full text
// and asserts, via the read-offset trace, that no octet of unit 2's or unit
// 3's segment range has been read at that point.
func TestFR_047_EmitsEarlierUnitTextBeforeReadingLaterUnitOctets(t *testing.T) {
	const segLen = 256
	r, ranges := buildContentDoc(t, 3, segLen)

	// decode returns the segment's first byte as a one-rune string, enough
	// to prove the unit's octets were read and delivered.
	decode := func(seg extract.ContentSegment, octets []byte) (string, error) {
		return string(rune(octets[0])), nil
	}
	ex, err := extract.NewExtractor(r, decode)
	if err != nil {
		t.Fatalf("NewExtractor: %v", err)
	}

	// Pull unit 1.
	u1, ok := ex.Next()
	if !ok {
		t.Fatalf("expected unit 1")
	}
	if u1.Text != string(rune(1)) {
		t.Fatalf("unit 1 text = %q, want segment-1 marker", u1.Text)
	}

	// At this point, no octet of unit 2's or unit 3's segment range may have
	// been read.
	for k := 1; k <= 2; k++ {
		if r.touched(ranges[k][0], ranges[k][1]) {
			t.Fatalf("after delivering unit 1, a read touched unit %d's segment range [%d,%d)", k+1, ranges[k][0], ranges[k][1])
		}
	}

	// Pulling unit 2 now reads unit 2's range (but still not unit 3's).
	u2, ok := ex.Next()
	if !ok || u2.Text != string(rune(2)) {
		t.Fatalf("unit 2 pull failed: ok=%v text=%q", ok, u2.Text)
	}
	if !r.touched(ranges[1][0], ranges[1][1]) {
		t.Fatalf("pulling unit 2 did not read its segment range")
	}
	if r.touched(ranges[2][0], ranges[2][1]) {
		t.Fatalf("pulling unit 2 wrongly read unit 3's segment range")
	}

	if u3, ok := ex.Next(); !ok || u3.Text != string(rune(3)) {
		t.Fatalf("unit 3 pull failed: ok=%v text=%q", ok, u3.Text)
	}
	if _, ok := ex.Next(); ok {
		t.Fatalf("expected exhaustion after 3 units")
	}
	if ex.Err() != nil {
		t.Fatalf("extractor error: %v", ex.Err())
	}
}
