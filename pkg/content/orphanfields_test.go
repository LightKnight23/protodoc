package content

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_030_OrphanRecordCapturesAllFourFields is T-0080's named test.
// Immediately after an orphaning delete, the resulting orphan record's
// Author, QuotedText, Prev and Next are all populated and match the
// pre-delete state's author and deleted text exactly (FR-030).
func TestFR_030_OrphanRecordCapturesAllFourFields(t *testing.T) {
	order := make(DocumentOrder, 5)
	for i := range order {
		id, err := MintID()
		if err != nil {
			t.Fatalf("MintID: %v", err)
		}
		order[i] = id
	}

	const preDeleteAuthor = uint16(4242)
	const anchoredText = "the quoted passage that is about to be deleted"

	ann := Annotation{
		Start: AnchorPoint{RunID: order[2]},
		End:   AnchorPoint{RunID: order[3]},
	}

	// Delete the annotation's whole anchored span (r2, r3). Surviving
	// neighbours: prev = r1, next = r4.
	deleted := map[pdlfmt.UnitID]struct{}{order[2]: {}, order[3]: {}}
	ctx := OrphanContext{Author: preDeleteAuthor, QuotedText: anchoredText}

	orphaned := OrphanOnDelete(ann, ctx, order, deleted)

	if !orphaned.Orphaned {
		t.Fatalf("annotation was not marked orphaned")
	}
	rec := orphaned.Orphan

	// Author matches the pre-delete author exactly.
	if rec.Author != preDeleteAuthor {
		t.Fatalf("orphan author = %d, want %d", rec.Author, preDeleteAuthor)
	}
	// QuotedText matches the deleted anchored text exactly.
	if rec.QuotedText != anchoredText {
		t.Fatalf("orphan quoted text = %q, want %q", rec.QuotedText, anchoredText)
	}
	// Prev and Next are populated with the resolved surviving neighbours.
	if !rec.Prev.Equal(order[1]) {
		t.Fatalf("orphan prev = %x, want r1", rec.Prev)
	}
	if !rec.Next.Equal(order[4]) {
		t.Fatalf("orphan next = %x, want r4", rec.Next)
	}

	// None of the four fields is left at its zero value in this case (all
	// four were genuinely captured, not silently lost).
	var zeroID pdlfmt.UnitID
	if rec.Author == 0 || rec.QuotedText == "" || rec.Prev.Equal(zeroID) || rec.Next.Equal(zeroID) {
		t.Fatalf("an orphan record field was left unpopulated: %+v", rec)
	}

	// The input annotation was not mutated.
	if ann.Orphaned {
		t.Fatalf("OrphanOnDelete mutated its input annotation")
	}
}
