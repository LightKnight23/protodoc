package content

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_028_FullDeletionOrphansRatherThanDrops is T-0078's named test.
// Deleting 100% of an annotation's anchored span leaves the annotation
// present in the document model, flagged orphaned, never silently dropped;
// an annotation with a surviving anchor endpoint stays live.
func TestFR_028_FullDeletionOrphansRatherThanDrops(t *testing.T) {
	runA, _ := MintID()
	runB, _ := MintID()
	runC, _ := MintID()
	annID1, _ := MintID()
	annID2, _ := MintID()

	// ann1 anchors entirely within runA (both endpoints on runA).
	ann1 := Annotation{
		ID:    annID1,
		Start: AnchorPoint{RunID: runA, BirthOrdinal: 0, Side: SideBefore, Boundary: BoundaryInside},
		End:   AnchorPoint{RunID: runA, BirthOrdinal: 5, Side: SideAfter, Boundary: BoundaryInside},
	}
	// ann2 spans runB..runC.
	ann2 := Annotation{
		ID:    annID2,
		Start: AnchorPoint{RunID: runB, BirthOrdinal: 0, Side: SideBefore, Boundary: BoundaryInside},
		End:   AnchorPoint{RunID: runC, BirthOrdinal: 3, Side: SideAfter, Boundary: BoundaryInside},
	}
	annotations := []Annotation{ann1, ann2}

	// Delete runA entirely (ann1's whole span) and runB (only ann2's start).
	deleted := map[pdlfmt.UnitID]struct{}{runA: {}, runB: {}}
	out := DeleteRange(annotations, deleted)

	// No annotation dropped: the slice still holds both.
	if len(out) != 2 {
		t.Fatalf("DeleteRange dropped annotations: got %d, want 2", len(out))
	}

	// ann1 (fully deleted span) is orphaned, not removed.
	if !out[0].Orphaned {
		t.Fatalf("ann1 whose entire anchored span was deleted was not orphaned")
	}
	if !out[0].ID.Equal(annID1) {
		t.Fatalf("ann1 identity changed under orphaning")
	}

	// ann2 (surviving end anchor on runC) stays live.
	if out[1].Orphaned {
		t.Fatalf("ann2 with a surviving anchor endpoint was wrongly orphaned")
	}

	// Deleting the remaining anchored run of ann2 now fully deletes its span
	// -> orphaned, still present.
	deleted2 := map[pdlfmt.UnitID]struct{}{runC: {}}
	// ann2's start already referenced runB (deleted); combine so both
	// endpoints are gone.
	deleted2[runB] = struct{}{}
	out2 := DeleteRange(out, deleted2)
	if len(out2) != 2 {
		t.Fatalf("second DeleteRange dropped annotations")
	}
	if !out2[1].Orphaned {
		t.Fatalf("ann2 not orphaned after its whole span was deleted")
	}

	// The input slice was not mutated by DeleteRange.
	if annotations[0].Orphaned || annotations[1].Orphaned {
		t.Fatalf("DeleteRange mutated its input slice")
	}
}
