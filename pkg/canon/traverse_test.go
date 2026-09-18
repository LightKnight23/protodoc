package canon

import (
	"bytes"
	"testing"

	"Protodoc/pkg/integrity"
	"Protodoc/pkg/pdlfmt"
)

func csUnit(seed byte, extra byte) pdlfmt.UnitID {
	var id pdlfmt.UnitID
	id[0] = seed
	id[15] = extra
	return id
}

// TestNFR_002_CanonicalTraversalOrderMatchesTC is T-0307's named unit test
// (NFR-002). The canon traversal order is byte-identical to T_C's own
// per-subtree canonical order (integrity.OrderRecords); the BOTTOM state yields
// zero records; two calls on the same state (and on any storage permutation)
// produce identical sequences.
func TestNFR_002_CanonicalTraversalOrderMatchesTC(t *testing.T) {
	// BOTTOM state: zero content records.
	if got := Traverse(&Document{}); len(got) != 0 {
		t.Errorf("BOTTOM state must yield zero records, got %d", len(got))
	}
	if got := Traverse(nil); got != nil {
		t.Errorf("nil state must yield nil, got %v", got)
	}

	// A set of subtrees in scrambled storage order.
	subtrees := []ContentSubtree{
		{UnitID: csUnit(0x30, 1), Frame: []byte("c")},
		{UnitID: csUnit(0x10, 2), Frame: []byte("a")},
		{UnitID: csUnit(0x20, 3), Frame: []byte("b")},
		{UnitID: csUnit(0x05, 4), Frame: []byte("z")},
	}
	state := &Document{Subtrees: subtrees}

	got := Traverse(state)

	// Cross-check: same order as integrity.OrderRecords over the same ids.
	recs := make([]integrity.ContentRecord, len(subtrees))
	for i, s := range subtrees {
		recs[i] = integrity.ContentRecord{UnitID: s.UnitID, Frame: s.Frame}
	}
	tcOrdered := integrity.OrderRecords(recs)
	if len(got) != len(tcOrdered) {
		t.Fatalf("lengths differ: canon %d, T_C %d", len(got), len(tcOrdered))
	}
	for i := range got {
		if got[i].UnitID != tcOrdered[i].UnitID {
			t.Errorf("order differs at %d: canon %x, T_C %x", i, got[i].UnitID[0], tcOrdered[i].UnitID[0])
		}
	}

	// Determinism + storage-order independence.
	got2 := Traverse(state)
	shuffled := &Document{Subtrees: []ContentSubtree{subtrees[2], subtrees[0], subtrees[3], subtrees[1]}}
	got3 := Traverse(shuffled)
	for i := range got {
		if got[i].UnitID != got2[i].UnitID || got[i].UnitID != got3[i].UnitID {
			t.Errorf("traversal not a pure function of state at %d", i)
		}
		if !bytes.Equal(got[i].Frame, got3[i].Frame) {
			t.Errorf("frame bytes differ under storage permutation at %d", i)
		}
	}

	// Input not mutated.
	if state.Subtrees[0].UnitID != csUnit(0x30, 1) {
		t.Errorf("Traverse mutated its input slice")
	}
}
