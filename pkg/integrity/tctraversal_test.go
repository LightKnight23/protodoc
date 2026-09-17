package integrity

import (
	"bytes"
	"math/rand"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_003_TCTraversalOrderDeterministicUnitIDLexicographic is T-0134's
// named test. OrderRecords returns records sorted by ascending unsigned
// big-endian byte-lexicographic order of their unit-ids, identically across
// two invocations on a shuffled input.
func TestFR_003_TCTraversalOrderDeterministicUnitIDLexicographic(t *testing.T) {
	const n = 200
	records := make([]ContentRecord, n)
	for i := range records {
		var id pdlfmt.UnitID
		// Spread ids across the byte space, embedding i so all are distinct.
		id[0] = byte(i * 37)
		id[1] = byte(i * 101)
		for k := 0; k < 8; k++ {
			id[8+k] = byte(uint64(i) >> (uint(k) * 8))
		}
		records[i] = ContentRecord{UnitID: id, Frame: []byte{byte(i)}}
	}

	// Order the original.
	ordered := OrderRecords(records)

	// The result is ascending byte-lexicographic by unit-id.
	for i := 1; i < len(ordered); i++ {
		if bytes.Compare(ordered[i-1].UnitID[:], ordered[i].UnitID[:]) >= 0 {
			t.Fatalf("records not in ascending unit-id order at index %d", i)
		}
	}

	// Shuffle a copy and order it: the ordinal assignment is identical.
	shuffled := append([]ContentRecord(nil), records...)
	r := rand.New(rand.NewSource(1234))
	r.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	orderedShuffled := OrderRecords(shuffled)

	if len(orderedShuffled) != len(ordered) {
		t.Fatalf("ordered lengths differ")
	}
	for i := range ordered {
		if ordered[i].UnitID != orderedShuffled[i].UnitID {
			t.Fatalf("ordinal assignment differs at %d between original and shuffled input", i)
		}
	}

	// The input slice is not mutated by OrderRecords.
	for i := range records {
		if records[i].UnitID != orderedInputCheck(i) {
			// records should still be in original construction order
			var id pdlfmt.UnitID
			id[0] = byte(i * 37)
			id[1] = byte(i * 101)
			for k := 0; k < 8; k++ {
				id[8+k] = byte(uint64(i) >> (uint(k) * 8))
			}
			if records[i].UnitID != id {
				t.Fatalf("OrderRecords mutated its input slice at %d", i)
			}
		}
	}
}

// orderedInputCheck reconstructs the original id at index i for the
// mutation check above.
func orderedInputCheck(i int) pdlfmt.UnitID {
	var id pdlfmt.UnitID
	id[0] = byte(i * 37)
	id[1] = byte(i * 101)
	for k := 0; k < 8; k++ {
		id[8+k] = byte(uint64(i) >> (uint(k) * 8))
	}
	return id
}
