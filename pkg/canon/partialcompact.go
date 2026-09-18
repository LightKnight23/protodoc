// Partial compaction (NFR-004; T-0312). Unlike full Compact, PartialCompact is
// PERMITTED while a signature is present. It is restricted to segments/ordinals
// named in NO present signature's covered_segment_ranges: it may relocate,
// coalesce and reclaim only UNCOVERED slots, and every covered segment's
// ordinal, offset, length and digest MUST remain byte-identical.
package canon

// SegmentSlot is a physical segment slot: its ordinal, offset, length and
// digest — the four fields that must stay byte-identical for covered segments.
type SegmentSlot struct {
	Ordinal uint16
	Offset  uint64
	Length  uint64
	Digest  [32]byte
}

// PartialCompactResult reports the compaction outcome: the resulting slots and
// which ordinals were reclaimed.
type PartialCompactResult struct {
	Slots             []SegmentSlot
	ReclaimedOrdinals []uint16
}

// PartialCompact compacts `slots`, reclaiming only ordinals NOT present in
// `covered` (the union of every present signature's covered_segment_ranges). A
// covered segment is left byte-identical (same ordinal/offset/length/digest);
// an uncovered segment may be dropped (reclaimed). It never refuses on account
// of a present signature — that is Compact's behavior, not this one.
func PartialCompact(slots []SegmentSlot, covered map[uint16]bool) PartialCompactResult {
	var res PartialCompactResult
	for _, s := range slots {
		if covered[s.Ordinal] {
			// Covered: preserved byte-identically (all four fields unchanged).
			res.Slots = append(res.Slots, s)
			continue
		}
		// Uncovered: reclaimed.
		res.ReclaimedOrdinals = append(res.ReclaimedOrdinals, s.Ordinal)
	}
	return res
}
