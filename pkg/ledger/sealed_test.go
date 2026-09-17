package ledger

import (
	"errors"
	"math/rand"
	"testing"
)

// TestFR_056_SealedSegmentRejectsMutation is T-0046's named test. It
// verifies the explicit write-once immutability guard both by direct API
// misuse (a write through the sealed segment's range is refused) and by a
// randomized, fuzz-style sequence of interleaved seal and write attempts
// (every write overlapping any sealed range is refused; every write clear
// of all sealed ranges is permitted).
func TestFR_056_SealedSegmentRejectsMutation(t *testing.T) {
	// --- Direct API-misuse case ---
	var g SealedSegmentGuard
	// Seal a segment at ordinal 0 spanning [PrefixLength, PrefixLength+256).
	base := uint64(PrefixLength)
	if err := g.Seal(0, base, 256); err != nil {
		t.Fatalf("Seal: %v", err)
	}

	// A write squarely inside the sealed range is refused.
	err := g.WriteRange(base+10, 20)
	if err == nil {
		t.Fatalf("write inside a sealed segment was not refused")
	}
	var immErr *SealImmutabilityError
	if !errors.As(err, &immErr) || !errors.Is(err, ErrSealedSegmentImmutable) {
		t.Fatalf("refusal error is %v (%T), want a SealImmutabilityError matching ErrSealedSegmentImmutable", err, err)
	}
	if immErr.Ordinal != 0 {
		t.Fatalf("refusal names ordinal %d, want 0", immErr.Ordinal)
	}

	// A write that only touches the very last octet of the sealed range is
	// still refused (boundary is inclusive of the sealed extent).
	if err := g.WriteRange(base+255, 1); err == nil {
		t.Fatalf("write touching the last sealed octet was not refused")
	}
	// A write straddling the sealed range's start is refused.
	if err := g.WriteRange(base-4, 8); err == nil {
		t.Fatalf("write straddling the sealed range start was not refused")
	}
	// An append past the sealed range is permitted.
	if err := g.WriteRange(base+256, 100); err != nil {
		t.Fatalf("append past the sealed range was wrongly refused: %v", err)
	}
	// A fixed-prefix patch (well before the sealed range) is permitted.
	if err := g.WriteRange(512, 48); err != nil {
		t.Fatalf("prefix-region write was wrongly refused: %v", err)
	}
	// Re-sealing the same ordinal is refused.
	if err := g.Seal(0, base, 256); !errors.Is(err, ErrSegmentAlreadySealed) {
		t.Fatalf("re-sealing ordinal 0 returned %v, want ErrSegmentAlreadySealed", err)
	}

	// --- Randomized interleaved seal/write sequence ---
	var rg SealedSegmentGuard
	r := rand.New(rand.NewSource(2026))
	// A model of the sealed ranges, checked independently of the guard.
	type rng struct{ start, end uint64 }
	var sealed []rng
	nextOffset := uint64(PrefixLength)
	var nextOrdinal uint64

	overlapsModel := func(s, e uint64) bool {
		for _, x := range sealed {
			if s < x.end && x.start < e {
				return true
			}
		}
		return false
	}

	for step := 0; step < 3000; step++ {
		if r.Intn(2) == 0 {
			// Seal a new segment at the current tail (never overlaps).
			length := uint64(r.Intn(500) + 1)
			if err := rg.Seal(nextOrdinal, nextOffset, length); err != nil {
				t.Fatalf("step %d: sealing a fresh tail segment failed: %v", step, err)
			}
			sealed = append(sealed, rng{nextOffset, nextOffset + length})
			nextOffset += length
			nextOrdinal++
		} else {
			// Attempt a write at a random offset within the used region.
			if nextOffset == uint64(PrefixLength) {
				continue // nothing sealed yet
			}
			span := nextOffset - uint64(PrefixLength)
			ws := uint64(PrefixLength) + uint64(r.Int63n(int64(span)+400))
			wl := uint64(r.Intn(300) + 1)
			gErr := rg.WriteRange(ws, wl)
			modelOverlap := overlapsModel(ws, ws+wl)
			if modelOverlap && gErr == nil {
				t.Fatalf("step %d: write [%d,%d) overlaps a sealed range but was permitted", step, ws, ws+wl)
			}
			if !modelOverlap && gErr != nil {
				t.Fatalf("step %d: write [%d,%d) clears all sealed ranges but was refused: %v", step, ws, ws+wl, gErr)
			}
			if modelOverlap && !errors.Is(gErr, ErrSealedSegmentImmutable) {
				t.Fatalf("step %d: overlapping write refused with %v, want ErrSealedSegmentImmutable", step, gErr)
			}
		}
	}
	if rg.SealedCount() == 0 {
		t.Fatalf("randomized sequence sealed no segments")
	}
}
