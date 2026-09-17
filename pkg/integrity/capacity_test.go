package integrity

import (
	"errors"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// mkSlots builds n populated CONTENT slots with distinct digests.
func mkSlots(n int) []container.SegmentTableSlot {
	slots := make([]container.SegmentTableSlot, n)
	for i := range slots {
		var d pdlfmt.Digest256
		d[0] = byte(i)
		d[1] = byte(i >> 8)
		d[2] = byte(i >> 16)
		slots[i] = container.SegmentTableSlot{SegmentType: container.SegmentTypeContent, Digest: d}
	}
	return slots
}

// TestTR_009_TSTreeAtLimitAndOnePastLimit is T-0144's named conformance test.
// TSRoot over MAX_SEGMENTS = 16384 populated leaves (well within the
// 16^4 = 65536 depth-4 capacity, exercising the headroom) returns a valid
// root with no error; TSRoot at exactly the T_S capacity (65536) also
// succeeds; and one past the tree builder's own capacity (65537) returns a
// named, distinguishable capacity error -- not a panic, not silent
// truncation. (The one-past-MAX_SEGMENTS=16385 case is a SegmentTable
// structural ceiling handled by M07's validator, not the tree builder,
// which has headroom to 65536.)
func TestTR_009_TSTreeAtLimitAndOnePastLimit(t *testing.T) {
	// At MAX_SEGMENTS: computes without error.
	atMax, err := TSRoot(mkSlots(container.MaxSegments)) // 16384
	if err != nil {
		t.Fatalf("TSRoot at MAX_SEGMENTS (16384): unexpected error %v", err)
	}
	var zero Digest
	if atMax == zero {
		t.Fatalf("TSRoot at MAX_SEGMENTS returned the zero digest")
	}

	// At exactly the T_S leaf capacity (65536): also computes without error.
	if _, err := TSRoot(mkSlots(TSLeafCapacity)); err != nil {
		t.Fatalf("TSRoot at T_S capacity (65536): unexpected error %v", err)
	}

	// One past the tree builder's capacity (65537): a named capacity error,
	// not a panic and not silent truncation.
	_, err = TSRoot(mkSlots(TSLeafCapacity + 1))
	if err == nil {
		t.Fatalf("TSRoot at 65537 did not error (silent truncation?)")
	}
	if !errors.Is(err, ErrTSTooManySegments) {
		t.Fatalf("TSRoot over-capacity error = %v, want ErrTSTooManySegments", err)
	}
}
