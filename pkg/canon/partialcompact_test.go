package canon

import "testing"

// TestNFR_004_PartialCompactionLeavesCoveredSegmentsByteIdentical is T-0312's
// named unit test (NFR-004). PartialCompact reclaims only uncovered ordinals; a
// field-by-field diff of every covered segment's slot before/after is empty;
// and it never refuses on account of a present signature.
func TestNFR_004_PartialCompactionLeavesCoveredSegmentsByteIdentical(t *testing.T) {
	mk := func(ord uint16, b byte) SegmentSlot {
		var d [32]byte
		d[0] = b
		return SegmentSlot{Ordinal: ord, Offset: uint64(ord) * 1024, Length: 512, Digest: d}
	}
	slots := []SegmentSlot{mk(0, 0xA), mk(1, 0xB), mk(2, 0xC), mk(3, 0xD)}
	// A SUBSET signature covers ordinals 0 and 2.
	covered := map[uint16]bool{0: true, 2: true}

	res := PartialCompact(slots, covered)

	// Only uncovered ordinals (1, 3) are reclaimed.
	got := map[uint16]bool{}
	for _, o := range res.ReclaimedOrdinals {
		got[o] = true
	}
	if len(got) != 2 || !got[1] || !got[3] {
		t.Errorf("reclaimed = %v, want exactly {1,3}", res.ReclaimedOrdinals)
	}

	// Every covered segment's slot is byte-identical before/after (field diff empty).
	before := map[uint16]SegmentSlot{0: mk(0, 0xA), 2: mk(2, 0xC)}
	seen := map[uint16]bool{}
	for _, s := range res.Slots {
		seen[s.Ordinal] = true
		want := before[s.Ordinal]
		if s != want {
			t.Errorf("covered ordinal %d changed: got %+v, want %+v", s.Ordinal, s, want)
		}
	}
	if !seen[0] || !seen[2] {
		t.Errorf("covered ordinals must be preserved, got %v", res.Slots)
	}

	// Never refuses with a present signature (covered set non-empty): it
	// returns a result, unlike Compact which would refuse.
	if len(res.Slots) == 0 && len(res.ReclaimedOrdinals) == 0 {
		t.Errorf("PartialCompact must proceed with a present signature")
	}
}
