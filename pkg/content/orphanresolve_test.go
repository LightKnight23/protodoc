package content

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_029_OrphanResolutionIsDeterministic is T-0079's named test.
// ResolveOrphan returns the same (prev, next) surviving-unit pair on every
// call for a fixed document state, including the boundary cases of orphaning
// at document start (no surviving predecessor) and document end (no
// surviving successor).
func TestFR_029_OrphanResolutionIsDeterministic(t *testing.T) {
	// Document order: r0 r1 r2 r3 r4 r5.
	order := make(DocumentOrder, 6)
	for i := range order {
		id, err := MintID()
		if err != nil {
			t.Fatalf("MintID: %v", err)
		}
		order[i] = id
	}
	var zero pdlfmt.UnitID

	mkAnn := func(start, end pdlfmt.UnitID) Annotation {
		return Annotation{
			Start:    AnchorPoint{RunID: start},
			End:      AnchorPoint{RunID: end},
			Orphaned: true,
		}
	}

	t.Run("middle span with surviving neighbours", func(t *testing.T) {
		// Orphan anchored on r2..r3; delete r2,r3. prev=r1, next=r4.
		deleted := map[pdlfmt.UnitID]struct{}{order[2]: {}, order[3]: {}}
		ann := mkAnn(order[2], order[3])
		for call := 0; call < 5; call++ {
			prev, next := ResolveOrphan(ann, order, deleted)
			if !prev.Equal(order[1]) || !next.Equal(order[4]) {
				t.Fatalf("call %d: got (prev,next)=(%x,%x), want (r1,r4)", call, prev, next)
			}
		}
	})

	t.Run("orphaned at document start", func(t *testing.T) {
		// Orphan anchored on r0..r1; delete r0,r1. No surviving predecessor
		// -> prev = zero; next = r2.
		deleted := map[pdlfmt.UnitID]struct{}{order[0]: {}, order[1]: {}}
		ann := mkAnn(order[0], order[1])
		prev, next := ResolveOrphan(ann, order, deleted)
		if !prev.Equal(zero) {
			t.Fatalf("start orphan prev = %x, want zero", prev)
		}
		if !next.Equal(order[2]) {
			t.Fatalf("start orphan next = %x, want r2", next)
		}
		// Deterministic across calls.
		p2, n2 := ResolveOrphan(ann, order, deleted)
		if p2 != prev || n2 != next {
			t.Fatalf("start orphan resolution not deterministic")
		}
	})

	t.Run("orphaned at document end", func(t *testing.T) {
		// Orphan anchored on r4..r5; delete r4,r5. prev=r3; no surviving
		// successor -> next = zero.
		deleted := map[pdlfmt.UnitID]struct{}{order[4]: {}, order[5]: {}}
		ann := mkAnn(order[4], order[5])
		prev, next := ResolveOrphan(ann, order, deleted)
		if !prev.Equal(order[3]) {
			t.Fatalf("end orphan prev = %x, want r3", prev)
		}
		if !next.Equal(zero) {
			t.Fatalf("end orphan next = %x, want zero", next)
		}
	})

	t.Run("entire document deleted", func(t *testing.T) {
		deleted := map[pdlfmt.UnitID]struct{}{}
		for _, id := range order {
			deleted[id] = struct{}{}
		}
		ann := mkAnn(order[2], order[3])
		prev, next := ResolveOrphan(ann, order, deleted)
		if !prev.Equal(zero) || !next.Equal(zero) {
			t.Fatalf("all-deleted resolution got (%x,%x), want (zero,zero)", prev, next)
		}
	})
}
