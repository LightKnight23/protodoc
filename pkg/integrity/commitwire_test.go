package integrity

import (
	"testing"

	"Protodoc/pkg/container"
)

// TestFR_003_CommitRingStateIDMatchesTCRoot is T-0137's named integration
// test. Two states differing in exactly one content record produce two
// distinct CommitRingRecord.state-id-field values, each equal to that
// state's independently-computed T_C_root; the record's own t-c-root field
// is set identically.
func TestFR_003_CommitRingStateIDMatchesTCRoot(t *testing.T) {
	stateA := []ContentRecord{mkRec(1, false), mkRec(2, true), mkRec(3, false)}
	// stateB differs in exactly one record's stored frame.
	stateB := append([]ContentRecord(nil), stateA...)
	stateB[1].Frame = append([]byte(nil), stateA[1].Frame...)
	stateB[1].Frame[0] ^= 0xFF

	var recA, recB container.CommitRingRecord
	rootA, err := SetCommitRingStateID(&recA, stateA)
	if err != nil {
		t.Fatalf("SetCommitRingStateID(A): %v", err)
	}
	rootB, err := SetCommitRingStateID(&recB, stateB)
	if err != nil {
		t.Fatalf("SetCommitRingStateID(B): %v", err)
	}

	// state-id-field equals the independently-computed T_C_root.
	indepA, _ := TCRoot(stateA)
	indepB, _ := TCRoot(stateB)
	if container.StateID(indepA) != recA.StateID {
		t.Fatalf("state A state-id-field != independent T_C_root")
	}
	if container.StateID(indepB) != recB.StateID {
		t.Fatalf("state B state-id-field != independent T_C_root")
	}

	// The two states differ in one record => differing state ids.
	if recA.StateID == recB.StateID {
		t.Fatalf("two states differing in one record produced the same state-id-field")
	}
	if rootA == rootB {
		t.Fatalf("returned T_C_roots equal for differing states")
	}

	// The record's own t-c-root field matches state-id-field.
	if [32]byte(recA.StateID) != recA.TCRoot {
		t.Fatalf("state A: t-c-root field != state-id-field")
	}
	if [32]byte(recB.StateID) != recB.TCRoot {
		t.Fatalf("state B: t-c-root field != state-id-field")
	}

	// Re-running on the identical state is stable (fresh recompute, not a
	// stale cache, but deterministic).
	var recA2 container.CommitRingRecord
	if _, err := SetCommitRingStateID(&recA2, stateA); err != nil {
		t.Fatalf("re-run A: %v", err)
	}
	if recA2.StateID != recA.StateID {
		t.Fatalf("state-id wiring not deterministic for the identical state")
	}
}
