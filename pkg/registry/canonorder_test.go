package registry

import (
	"math/rand"
	"testing"

	"Protodoc/pkg/content"
)

// TestFR_012_ExtPositionKeyCanonicalOrder is T-0058's named unit test
// (FR-012 / DP-011). It asserts the canonical order among extension envelopes
// at one position is by (ext-tok, ext-position-key), byte-lexicographic, and
// that the ordering is total and deterministic (a shuffled input always sorts
// to the identical sequence).
func TestFR_012_ExtPositionKeyCanonicalOrder(t *testing.T) {
	mk := func(tokOwner, tokSeq uint32, apSeed byte, ordinal uint32) ExtEnvelope {
		return ExtEnvelope{
			Tok:         NewExtToken(tokOwner, tokSeq),
			Disposition: DispositionRefuse,
			FallbackRef: Zero16,
			PositionKey: content.AnchorPoint{
				RunID:        fixtureUnitID(apSeed),
				BirthOrdinal: ordinal,
				Side:         content.SideBefore,
				Boundary:     content.BoundaryInside,
			},
		}
	}

	// Canonical expected order: by ext-tok bytes first, then position key.
	// tok (1,1) < (1,2) < (2,0); within (1,1), position key ordinal 5 < 9.
	want := []ExtEnvelope{
		mk(1, 1, 0x10, 5),
		mk(1, 1, 0x10, 9),
		mk(1, 2, 0x10, 0),
		mk(2, 0, 0x10, 0),
	}

	// Sanity: the expected order is strictly ascending under the comparator.
	for i := 1; i < len(want); i++ {
		if CompareEnvelopeCanonical(want[i-1], want[i]) >= 0 {
			t.Fatalf("expected-order element %d does not strictly follow %d", i, i-1)
		}
	}

	// Shuffle many times; each sort must reproduce `want` exactly.
	rng := rand.New(rand.NewSource(1))
	for trial := 0; trial < 50; trial++ {
		shuffled := append([]ExtEnvelope(nil), want...)
		rng.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
		SortEnvelopesCanonical(shuffled)
		for i := range want {
			if CompareEnvelopeCanonical(shuffled[i], want[i]) != 0 {
				t.Fatalf("trial %d: sorted[%d] != want[%d] (tok %x vs %x)", trial, i, i, shuffled[i].Tok, want[i].Tok)
			}
		}
	}

	// The comparator orders by ext-tok BEFORE position key: a smaller
	// position key with a larger tok still sorts after a larger position key
	// with a smaller tok.
	lowTokHighPos := mk(1, 1, 0xFF, 9999)
	highTokLowPos := mk(2, 0, 0x00, 0)
	if CompareEnvelopeCanonical(lowTokHighPos, highTokLowPos) >= 0 {
		t.Error("ext-tok must dominate the ordering over the position key")
	}

	// Equal tok and position key compare equal (0).
	e := mk(3, 3, 0x22, 7)
	if CompareEnvelopeCanonical(e, e) != 0 {
		t.Error("identical envelopes must compare equal")
	}
}
