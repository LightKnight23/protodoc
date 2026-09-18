package integrity

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// tcSeqUnit builds a distinct unit-id seeded by b.
func tcSeqUnit(b byte) pdlfmt.UnitID {
	var id pdlfmt.UnitID
	id[0] = b
	id[15] = b ^ 0x5a
	return id
}

// TestFR_036_TCTraversalUsesRootSequence is T-0275's named integration test
// (FR-036). It verifies the T_C traversal is keyed on ROOT_SEQUENCE reading
// order — not the interim CSPRNG unit-id order — so that T_C_root changes when
// and only when ROOT_SEQUENCE changes: permuting storage order leaves the root
// identical, while reordering rs-order changes it.
func TestFR_036_TCTraversalUsesRootSequence(t *testing.T) {
	uA, uB, uC := tcSeqUnit(0x10), tcSeqUnit(0x20), tcSeqUnit(0x30)
	recs := []ContentRecord{
		{UnitID: uA, Frame: []byte("alpha")},
		{UnitID: uB, Frame: []byte("bravo")},
		{UnitID: uC, Frame: []byte("charlie")},
	}
	// The authored reading order.
	order := []pdlfmt.UnitID{uA, uB, uC}

	root1, err := TCRootWithSequence(recs, order)
	if err != nil {
		t.Fatalf("TCRootWithSequence: %v", err)
	}

	// (1) Permuting the STORAGE order of the records must NOT change the root,
	// because the traversal is keyed on ROOT_SEQUENCE, not storage order.
	permuted := []ContentRecord{recs[2], recs[0], recs[1]}
	root2, err := TCRootWithSequence(permuted, order)
	if err != nil {
		t.Fatalf("TCRootWithSequence permuted: %v", err)
	}
	if root1 != root2 {
		t.Errorf("T_C root changed on a mere storage-order permutation; traversal is not keyed purely on ROOT_SEQUENCE")
	}

	// (2) Reordering ROOT_SEQUENCE (rs-order) MUST change the root.
	reordered := []pdlfmt.UnitID{uC, uA, uB}
	root3, err := TCRootWithSequence(recs, reordered)
	if err != nil {
		t.Fatalf("TCRootWithSequence reordered: %v", err)
	}
	if root1 == root3 {
		t.Errorf("T_C root did not change when ROOT_SEQUENCE changed; traversal does not track reading order")
	}

	// (3) The sequence-keyed order differs from the interim CSPRNG (unit-id
	// lexicographic) order, proving the retrofit actually replaced the interim
	// rule rather than coinciding with it. Use a reading order that is NOT the
	// lexicographic order of the unit-ids.
	seqKeyed := OrderRecordsBySequence(recs, reordered)
	interim := OrderRecords(recs)
	differs := false
	for i := range seqKeyed {
		if seqKeyed[i].UnitID != interim[i].UnitID {
			differs = true
			break
		}
	}
	if !differs {
		t.Errorf("sequence-keyed order coincides with the interim CSPRNG order; retrofit not demonstrated")
	}
}
