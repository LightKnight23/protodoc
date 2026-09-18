package history

import (
	"errors"
	"testing"
)

// TestFR_005_NearestCommonAncestorOffline is T-0211's named integration test
// (FR-005). Two divergent copies each carry their own operation log with a
// shared prefix; the nearest common ancestor of the two divergent tip states
// is computed from the two logs alone -- no coordinating service, no network.
func TestFR_005_NearestCommonAncestorOffline(t *testing.T) {
	// Shared prefix: s1 (root) <- s2 <- s3.
	s1, s2, s3 := hStateID(0x01), hStateID(0x02), hStateID(0x03)
	shared := []OperationRecord{
		{Kind: OpValueClaim, Target: hUnitID(1), OrderKey: s2, Predecessor: s1},
		{Kind: OpValueClaim, Target: hUnitID(1), OrderKey: s3, Predecessor: s2},
	}

	// Copy A diverges from s3: s3 <- a1 <- a2. Its file's log = shared + A ops.
	a1, a2 := hStateID(0xA1), hStateID(0xA2)
	logA := append(append([]OperationRecord(nil), shared...),
		OperationRecord{Kind: OpValueClaim, Target: hUnitID(1), OrderKey: a1, Predecessor: s3},
		OperationRecord{Kind: OpValueClaim, Target: hUnitID(1), OrderKey: a2, Predecessor: a1},
	)

	// Copy B diverges from s3: s3 <- b1 <- b2. Its file's log = shared + B ops.
	b1, b2 := hStateID(0xB1), hStateID(0xB2)
	logB := append(append([]OperationRecord(nil), shared...),
		OperationRecord{Kind: OpValueClaim, Target: hUnitID(1), OrderKey: b1, Predecessor: s3},
		OperationRecord{Kind: OpValueClaim, Target: hUnitID(1), OrderKey: b2, Predecessor: b1},
	)

	// Determine the NCA from the two files alone: build each file's lineage,
	// merge them (the shared prefix's state-ids match in both), and query.
	lin := NewLineage(logA)
	linB := NewLineage(logB)
	lin.Merge(linB)

	nca, err := lin.NearestCommonAncestor(a2, b2)
	if err != nil {
		t.Fatalf("NearestCommonAncestor(a2, b2): %v", err)
	}
	if nca != s3 {
		t.Fatalf("nearest common ancestor = %x, want s3 %x", nca, s3)
	}

	// Symmetric-ish: NCA of a2 and a state on the shared prefix (s2) is s2.
	if got, err := lin.NearestCommonAncestor(a2, s2); err != nil || got != s2 {
		t.Errorf("NCA(a2, s2) = %x err %v, want s2", got, err)
	}

	// NCA of a tip and the root is the root.
	if got, err := lin.NearestCommonAncestor(b2, s1); err != nil || got != s1 {
		t.Errorf("NCA(b2, s1) = %x err %v, want s1 (root)", got, err)
	}

	// Two unrelated states (no shared ancestor) error.
	unrelated := NewLineage([]OperationRecord{
		{Kind: OpValueClaim, Target: hUnitID(2), OrderKey: hStateID(0xC1), Predecessor: hStateID(0xC0)},
	})
	lin.Merge(unrelated)
	if _, err := lin.NearestCommonAncestor(a2, hStateID(0xC1)); !errors.Is(err, ErrNoCommonAncestor) {
		t.Errorf("unrelated states: err = %v, want ErrNoCommonAncestor", err)
	}
}
