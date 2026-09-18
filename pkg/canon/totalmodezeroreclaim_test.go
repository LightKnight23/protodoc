package canon

import "testing"

// TestCONF_COMPACT_002_TotalModeZeroReclaim is T-0313's named conformance test
// (vector CONF-COMPACT-002; NFR-004, plan.md S8 residual-risk note). A
// TOTAL-mode-signed document — where the signature covers EVERY segment — has
// zero uncovered ordinals, so PartialCompact() is a well-defined no-op
// (reclaims nothing, changes nothing), distinct from a SUBSET-signed document
// where uncovered slots exist to reclaim.
func TestCONF_COMPACT_002_TotalModeZeroReclaim(t *testing.T) {
	mk := func(ord uint16) SegmentSlot {
		var d [32]byte
		d[0] = byte(ord)
		return SegmentSlot{Ordinal: ord, Offset: uint64(ord) * 1024, Length: 256, Digest: d}
	}
	slots := []SegmentSlot{mk(0), mk(1), mk(2), mk(3)}

	// TOTAL mode: the signature covers EVERY ordinal.
	total := map[uint16]bool{0: true, 1: true, 2: true, 3: true}
	res := PartialCompact(slots, total)

	// No-op: nothing reclaimed, every slot preserved byte-identically.
	if len(res.ReclaimedOrdinals) != 0 {
		t.Errorf("TOTAL-mode PartialCompact must reclaim nothing, reclaimed %v", res.ReclaimedOrdinals)
	}
	if len(res.Slots) != len(slots) {
		t.Fatalf("TOTAL-mode PartialCompact must preserve all %d slots, got %d", len(slots), len(res.Slots))
	}
	for i := range slots {
		if res.Slots[i] != slots[i] {
			t.Errorf("slot %d changed under TOTAL-mode no-op: got %+v, want %+v", i, res.Slots[i], slots[i])
		}
	}

	// Contrast: a SUBSET-signed document (ordinal 3 uncovered) does reclaim.
	subset := map[uint16]bool{0: true, 1: true, 2: true}
	if r := PartialCompact(slots, subset); len(r.ReclaimedOrdinals) != 1 || r.ReclaimedOrdinals[0] != 3 {
		t.Errorf("SUBSET mode should reclaim ordinal 3, got %v", r.ReclaimedOrdinals)
	}
}
