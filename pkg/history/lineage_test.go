package history

import (
	"errors"
	"testing"

	"Protodoc/pkg/container"
)

// TestFR_005_PredecessorChainWalkable is T-0210's named unit test (FR-005).
// The op-predecessor fields form a causal chain; from a state's operation log
// the predecessor chain back to the root is walkable, deterministically and
// without any external service.
func TestFR_005_PredecessorChainWalkable(t *testing.T) {
	// A linear history: s1 (root) <- s2 <- s3 <- s4.
	s1, s2, s3, s4 := hStateID(0x01), hStateID(0x02), hStateID(0x03), hStateID(0x04)
	ops := []OperationRecord{
		{Kind: OpValueClaim, Target: hUnitID(1), OrderKey: s2, Predecessor: s1},
		{Kind: OpValueClaim, Target: hUnitID(1), OrderKey: s3, Predecessor: s2},
		{Kind: OpValueClaim, Target: hUnitID(1), OrderKey: s4, Predecessor: s3},
	}
	l := NewLineage(ops)

	chain, err := l.WalkPredecessors(s4)
	if err != nil {
		t.Fatalf("WalkPredecessors(s4): %v", err)
	}
	want := []container.StateID{s4, s3, s2, s1}
	if len(chain) != len(want) {
		t.Fatalf("chain length %d, want %d: %v", len(chain), len(want), chain)
	}
	for i := range want {
		if chain[i] != want[i] {
			t.Errorf("chain[%d] = %x, want %x", i, chain[i], want[i])
		}
	}

	// Walking from a mid-state stops at the root.
	mid, err := l.WalkPredecessors(s2)
	if err != nil {
		t.Fatalf("WalkPredecessors(s2): %v", err)
	}
	if len(mid) != 2 || mid[0] != s2 || mid[1] != s1 {
		t.Errorf("mid-chain = %v, want [s2 s1]", mid)
	}

	// An unknown state errors.
	if _, err := l.WalkPredecessors(hStateID(0x99)); !errors.Is(err, ErrUnknownState) {
		t.Errorf("unknown state: err = %v, want ErrUnknownState", err)
	}

	// A cycle (malformed log) is bounded, not an infinite loop.
	cyc := Lineage{}
	a, b := hStateID(0xAA), hStateID(0xBB)
	cyc.AddState(a, b)
	cyc.AddState(b, a)
	c, err := cyc.WalkPredecessors(a)
	if err != nil {
		t.Fatalf("cyclic walk errored: %v", err)
	}
	if len(c) > 2 {
		t.Errorf("cyclic walk did not terminate cleanly: %v", c)
	}

	// Determinism: rebuilding from a reordered log yields the same chain.
	reordered := []OperationRecord{ops[2], ops[0], ops[1]}
	chain2, _ := NewLineage(reordered).WalkPredecessors(s4)
	if len(chain2) != len(want) {
		t.Errorf("reordered-log chain length %d, want %d", len(chain2), len(want))
	}
}
