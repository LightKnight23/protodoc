package validate

import (
	"errors"
	"testing"

	"Protodoc/pkg/ceilings"
)

// guardedAlloc models an allocation site: it guards the declared count first
// and only allocates (recording the allocation size) if the guard passes.
// The test uses it to prove no oversized allocation happens on the failing
// path.
func guardedAlloc(ceilingName string, declared uint64, allocSizes *[]uint64) error {
	if err := GuardCount(ceilingName, declared); err != nil {
		return err // aborts BEFORE any allocation
	}
	*allocSizes = append(*allocSizes, declared)
	return nil
}

// TestFR_106_CeilingPreAllocationAbortsBeforeAllocation is T-0115's named
// conformance test. For each M07-guarded CON-009 ceiling, an at-limit value
// passes and allocates, and a one-past-limit value aborts naming the ceiling
// before any allocation is recorded.
func TestFR_106_CeilingPreAllocationAbortsBeforeAllocation(t *testing.T) {
	// The M07-owned count ceilings (the frame-directory boundary is M01's).
	names := []string{
		"SegmentTable slot count (MAX_SEGMENTS)",
		"MAX_FRAMES_PER_SEGMENT",
		"MAX_DECODED_UNIT",
		"MAX_TEXT_UNIT_OCTETS",
		"MAX_CONTENT_UNITS",
		"MAX_SIGNATURES",
	}
	for _, name := range names {
		max := ceilings.MustMax(name)

		// At-limit: passes and allocates.
		var allocs []uint64
		if err := guardedAlloc(name, max, &allocs); err != nil {
			t.Fatalf("%s: at-limit (%d) rejected: %v", name, max, err)
		}
		if len(allocs) != 1 || allocs[0] != max {
			t.Fatalf("%s: at-limit did not allocate exactly the limit", name)
		}

		// One past limit: aborts before allocation.
		allocs = nil
		err := guardedAlloc(name, max+1, &allocs)
		if err == nil {
			t.Fatalf("%s: one-past-limit (%d) was not rejected", name, max+1)
		}
		if !errors.Is(err, ErrCeilingExceeded) {
			t.Fatalf("%s: one-past-limit error = %v, want ErrCeilingExceeded", name, err)
		}
		var ce *CeilingError
		if !errors.As(err, &ce) || ce.CeilingName != name || ce.Observed != max+1 {
			t.Fatalf("%s: CeilingError = %+v, want name=%s observed=%d", name, ce, name, max+1)
		}
		if len(allocs) != 0 {
			t.Fatalf("%s: an allocation occurred on the failing path (want zero, abort-before-allocate)", name)
		}
	}

	// GuardAllocation's overflow check: a count within the ceiling but whose
	// product overflows is rejected before allocation.
	// Use MAX_DECODED_UNIT (large) with a huge element size.
	if err := GuardAllocation("MAX_DECODED_UNIT", ceilings.MustMax("MAX_DECODED_UNIT"), ^uint64(0)); err == nil {
		t.Fatalf("GuardAllocation did not reject a product overflow")
	}
}
