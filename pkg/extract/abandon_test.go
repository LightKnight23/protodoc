package extract_test

import (
	"testing"

	"Protodoc/pkg/extract"
)

// TestFR_048_AbandoningExtractionReadsNoFurtherOctets is T-0095's named
// integration test. Cancelling extraction after unit 2 of a 5-unit fixture
// results in zero additional ReadAt calls beyond what units 1-2 required,
// verified by a read-count assertion on the instrumented reader.
func TestFR_048_AbandoningExtractionReadsNoFurtherOctets(t *testing.T) {
	const segLen = 256
	r, _ := buildContentDoc(t, 5, segLen)

	decode := func(seg extract.ContentSegment, octets []byte) (string, error) {
		return string(rune(octets[0])), nil
	}
	ex, err := extract.NewExtractor(r, decode)
	if err != nil {
		t.Fatalf("NewExtractor: %v", err)
	}

	// Consume units 1 and 2.
	if _, ok := ex.Next(); !ok {
		t.Fatalf("unit 1 pull failed")
	}
	if _, ok := ex.Next(); !ok {
		t.Fatalf("unit 2 pull failed")
	}
	readsAfterTwoUnits := len(r.reads)

	// Abandon: Stop, then any further Next must read nothing.
	ex.Stop()
	if _, ok := ex.Next(); ok {
		t.Fatalf("Next after Stop returned a unit; extraction was not abandoned")
	}
	if _, ok := ex.Next(); ok {
		t.Fatalf("second Next after Stop returned a unit")
	}

	// Zero additional ReadAt calls occurred after abandonment.
	if len(r.reads) != readsAfterTwoUnits {
		t.Fatalf("abandonment performed %d further reads, want 0", len(r.reads)-readsAfterTwoUnits)
	}
}

// TestFR_048_BreakingLoopReadsNoFurtherOctets confirms the same guarantee
// when a caller simply stops pulling (breaks its loop) after unit 2 without
// calling Stop: no read happens for units it never pulled.
func TestFR_048_BreakingLoopReadsNoFurtherOctets(t *testing.T) {
	const segLen = 256
	r, ranges := buildContentDoc(t, 5, segLen)
	decode := func(seg extract.ContentSegment, octets []byte) (string, error) { return "x", nil }
	ex, err := extract.NewExtractor(r, decode)
	if err != nil {
		t.Fatalf("NewExtractor: %v", err)
	}
	ex.Next()
	ex.Next()
	// Caller breaks here, never pulling units 3-5. Their segment ranges must
	// be untouched.
	for k := 2; k < 5; k++ {
		if r.touched(ranges[k][0], ranges[k][1]) {
			t.Fatalf("unit %d segment range was read despite the caller stopping after unit 2", k+1)
		}
	}
}
