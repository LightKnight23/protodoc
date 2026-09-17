package content

import (
	"math/rand"
	"testing"
)

// TestFR_020_MergeIsPureSyntacticPredicate is T-0070's named test. It
// asserts CanMergeRuns is true exactly when the two runs share run_id and
// are base_ordinal-contiguous, and that MergeRuns is the exact left-inverse
// of SplitRun for every generated split point.
func TestFR_020_MergeIsPureSyntacticPredicate(t *testing.T) {
	r := rand.New(rand.NewSource(11))

	// --- Merge is the exact left-inverse of split ---
	for trial := 0; trial < 500; trial++ {
		rid, err := testMintID()
		if err != nil {
			t.Fatalf("MintID: %v", err)
		}
		run := Run{RunID: rid, BaseOrdinal: uint32(r.Intn(1000)), Text: randomRunText(r, r.Intn(40))}
		for at := 0; at <= run.ScalarLen(); at++ {
			left, right, ok := SplitRun(run, at)
			if !ok {
				t.Fatalf("SplitRun at %d failed", at)
			}
			if !CanMergeRuns(left, right) {
				t.Fatalf("trial %d at %d: split outputs are not mergeable", trial, at)
			}
			merged, ok := MergeRuns(left, right)
			if !ok {
				t.Fatalf("trial %d at %d: MergeRuns of split outputs failed", trial, at)
			}
			// Exact reconstruction of the original run.
			if !merged.RunID.Equal(run.RunID) || merged.BaseOrdinal != run.BaseOrdinal || merged.Text != run.Text {
				t.Fatalf("trial %d at %d: MergeRuns(SplitRun(r)) != r", trial, at)
			}
		}
	}

	// --- Predicate is exactly run_id-match AND base_ordinal-contiguity ---
	ridA, _ := testMintID()
	ridB, _ := testMintID()
	a := Run{RunID: ridA, BaseOrdinal: 10, Text: "abc"} // EndOrdinal 13

	// Contiguous, same run_id: mergeable.
	if !CanMergeRuns(a, Run{RunID: ridA, BaseOrdinal: 13, Text: "de"}) {
		t.Fatalf("contiguous same-lineage runs reported not mergeable")
	}
	// Same run_id but a GAP in base_ordinal: not mergeable.
	if CanMergeRuns(a, Run{RunID: ridA, BaseOrdinal: 14, Text: "de"}) {
		t.Fatalf("non-contiguous runs reported mergeable")
	}
	// Same run_id but OVERLAP: not mergeable.
	if CanMergeRuns(a, Run{RunID: ridA, BaseOrdinal: 12, Text: "de"}) {
		t.Fatalf("overlapping runs reported mergeable")
	}
	// Contiguous but DIFFERENT run_id: not mergeable (distinct lineage).
	if CanMergeRuns(a, Run{RunID: ridB, BaseOrdinal: 13, Text: "de"}) {
		t.Fatalf("different-lineage runs reported mergeable")
	}
	// MergeRuns refuses exactly when CanMergeRuns is false.
	if _, ok := MergeRuns(a, Run{RunID: ridB, BaseOrdinal: 13, Text: "de"}); ok {
		t.Fatalf("MergeRuns merged different-lineage runs")
	}

	// --- Purity: merge depends only on run_id and base_ordinal, not on
	// text content. Two runs with identical (run_id, contiguous ordinals)
	// merge regardless of what their text says. ---
	c1 := Run{RunID: ridA, BaseOrdinal: 0, Text: "\U0001F600"} // 1 scalar
	c2 := Run{RunID: ridA, BaseOrdinal: 1, Text: "xyz"}
	if !CanMergeRuns(c1, c2) {
		t.Fatalf("merge predicate is not purely (run_id, ordinal): identical-lineage contiguous runs rejected")
	}
}
