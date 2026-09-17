package content

import (
	"sort"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// makeRunSeq builds n runs with distinct minted run_ids and marker text, so
// each run is individually identifiable after a move or reorder.
func makeRunSeq(t *testing.T, n int) []Run {
	t.Helper()
	runs := make([]Run, n)
	for i := range runs {
		rid, err := testMintID()
		if err != nil {
			t.Fatalf("MintID: %v", err)
		}
		runs[i] = Run{RunID: rid, BaseOrdinal: uint32(i * 10), Text: string(rune('A' + i))}
	}
	return runs
}

// idMultiset returns the sorted run_id octet strings of a run sequence, for
// comparing that the SET of identities is preserved regardless of order.
func idMultiset(runs []Run) []string {
	ids := make([]string, len(runs))
	for i, r := range runs {
		ids[i] = string(r.RunID[:])
	}
	sort.Strings(ids)
	return ids
}

// TestFR_020_MoveReorderPreservesRunID is T-0071's named test. It confirms
// MoveRun and ReorderRuns change only position, never run_id (or
// base_ordinal or text), across single-move, multi-move and cyclic-reorder
// cases in a table-driven test.
func TestFR_020_MoveReorderPreservesRunID(t *testing.T) {
	// --- MoveRun cases (single and repeated moves) ---
	moveCases := []struct {
		name string
		n    int
		from []int
		to   []int
	}{
		{"single move forward", 5, []int{1}, []int{3}},
		{"single move backward", 5, []int{4}, []int{0}},
		{"multi move", 6, []int{0, 5, 2}, []int{3, 0, 4}},
	}
	for _, c := range moveCases {
		t.Run(c.name, func(t *testing.T) {
			runs := makeRunSeq(t, c.n)
			origIDs := idMultiset(runs)
			// Track each id's associated text to prove content rides with id.
			textByID := map[pdlfmt.UnitID]string{}
			for _, r := range runs {
				textByID[r.RunID] = r.Text
			}

			cur := runs
			for i := range c.from {
				var ok bool
				cur, ok = MoveRun(cur, c.from[i], c.to[i])
				if !ok {
					t.Fatalf("MoveRun(%d,%d) not ok", c.from[i], c.to[i])
				}
			}
			if len(cur) != c.n {
				t.Fatalf("after moves length %d, want %d (a run was lost or duplicated)", len(cur), c.n)
			}
			// The multiset of run_ids is unchanged.
			for i, id := range idMultiset(cur) {
				if id != origIDs[i] {
					t.Fatalf("move changed the set of run_ids")
				}
			}
			// Each run's text still matches the text its id had originally
			// (identity and content rode together; only order changed).
			for _, r := range cur {
				if textByID[r.RunID] != r.Text {
					t.Fatalf("run %x text changed under move: %q vs %q", r.RunID, r.Text, textByID[r.RunID])
				}
			}
		})
	}

	// --- ReorderRuns: cyclic and reverse permutations ---
	reorderCases := []struct {
		name string
		perm []int
	}{
		{"reverse", []int{4, 3, 2, 1, 0}},
		{"cyclic shift", []int{1, 2, 3, 4, 0}},
		{"identity", []int{0, 1, 2, 3, 4}},
	}
	for _, c := range reorderCases {
		t.Run("reorder "+c.name, func(t *testing.T) {
			runs := makeRunSeq(t, 5)
			origIDs := idMultiset(runs)
			textByID := map[pdlfmt.UnitID]string{}
			for _, r := range runs {
				textByID[r.RunID] = r.Text
			}

			out, ok := ReorderRuns(runs, c.perm)
			if !ok {
				t.Fatalf("ReorderRuns not ok for perm %v", c.perm)
			}
			// out[i] must be the run originally at perm[i], run_id intact.
			for i, p := range c.perm {
				if !out[i].RunID.Equal(runs[p].RunID) {
					t.Fatalf("reorder: out[%d] run_id != runs[%d] run_id", i, p)
				}
				if out[i].Text != runs[p].Text || out[i].BaseOrdinal != runs[p].BaseOrdinal {
					t.Fatalf("reorder: out[%d] content/base_ordinal changed", i)
				}
			}
			for i, id := range idMultiset(out) {
				if id != origIDs[i] {
					t.Fatalf("reorder changed the set of run_ids")
				}
			}
			_ = textByID
		})
	}

	// Invalid permutations and out-of-range moves are rejected.
	runs := makeRunSeq(t, 3)
	if _, ok := ReorderRuns(runs, []int{0, 0, 1}); ok {
		t.Fatalf("ReorderRuns accepted a non-permutation")
	}
	if _, ok := MoveRun(runs, 5, 0); ok {
		t.Fatalf("MoveRun accepted an out-of-range index")
	}
}
