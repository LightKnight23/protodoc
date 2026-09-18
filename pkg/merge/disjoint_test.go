package merge

import (
	"bytes"
	"math/rand"
	"testing"

	"Protodoc/pkg/history"
	"Protodoc/pkg/pdlfmt"
)

func mTarget(seed byte) pdlfmt.UnitID {
	var id pdlfmt.UnitID
	for i := range id {
		id[i] = seed + byte(i)
	}
	return id
}

// TestFR_092_DisjointCommuteOrderIndependent is T-0226's named unit test
// (FR-092). Disjoint concurrent operations commute: given the same set of
// changes in ANY delivery order, the merged document state is identical. Every
// disjoint pair classifies as DISJOINT-COMMUTE, and applying them in any
// permutation yields the same canonical state.
func TestFR_092_DisjointCommuteOrderIndependent(t *testing.T) {
	// Six operations, each on a DISTINCT target (pairwise disjoint).
	ops := []Op{
		{Kind: history.OpValueClaim, Target: mTarget(0x01), Value: []byte("a")},
		{Kind: history.OpSequencePositionClaim, Target: mTarget(0x11), Value: []byte("b")},
		{Kind: history.OpValueClaim, Target: mTarget(0x21), Value: []byte("c")},
		{Kind: history.OpDelete, Target: mTarget(0x31)},
		{Kind: history.OpValueClaim, Target: mTarget(0x41), Value: []byte("e")},
		{Kind: history.OpSequencePositionClaim, Target: mTarget(0x51), Value: []byte("f")},
	}
	base := Doc{mTarget(0x31): []byte("to be deleted")}

	// Every disjoint pair classifies as DISJOINT-COMMUTE.
	for i := range ops {
		for j := range ops {
			if i == j {
				continue
			}
			o, err := Classify(ops[i].Kind, ops[j].Kind, Disjoint)
			if err != nil || o != DisjointCommute {
				t.Errorf("disjoint pair (%v,%v) classified %v err %v, want DISJOINT-COMMUTE", ops[i].Kind, ops[j].Kind, o, err)
			}
		}
	}

	// Applying in the given order yields a reference state.
	reference := ApplyDisjoint(base, ops).canonical()

	// Applying in many random permutations yields the identical canonical
	// state -- order-independent.
	rng := rand.New(rand.NewSource(7))
	for trial := 0; trial < 100; trial++ {
		perm := append([]Op(nil), ops...)
		rng.Shuffle(len(perm), func(i, j int) { perm[i], perm[j] = perm[j], perm[i] })
		got := ApplyDisjoint(base, perm).canonical()
		if !bytes.Equal(got, reference) {
			t.Fatalf("trial %d: disjoint apply is order-dependent", trial)
		}
	}
}
