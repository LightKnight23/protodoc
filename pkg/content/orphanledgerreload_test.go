package content_test

import (
	"testing"

	"Protodoc/pkg/content"
	"Protodoc/pkg/ledger"
	"Protodoc/pkg/pdlfmt"
)

// TestM04_OrphanCarriageSurvivesLedgerReload is T-0086's named conformance
// test (case id CONF-M04-ORPHAN-CARRIAGE-LEDGER-RELOAD): the M04 exit
// criterion. It orphans an annotation, saves it by appending an
// orphaned-annotation segment via the append-only ledger, reloads the
// segment from the resulting image, and confirms all four FR-030 orphan
// fields (author, quoted text, prev, next) plus both FR-026/FR-027
// boundary-behaviour values are bit-identical to the pre-save state
// (FR-027, FR-028, FR-029, FR-030).
func TestM04_OrphanCarriageSurvivesLedgerReload(t *testing.T) {
	// Build a document order and orphan an annotation whose whole span is
	// deleted, capturing all four fields via the M04 orphaning path.
	order := make(content.DocumentOrder, 5)
	for i := range order {
		id, err := content.MintID()
		if err != nil {
			t.Fatalf("MintID: %v", err)
		}
		order[i] = id
	}
	annID, _ := content.MintID()
	bodyBlock, _ := content.MintID()

	ann := content.Annotation{
		ID:        annID,
		Start:     content.AnchorPoint{RunID: order[2], BirthOrdinal: 3, Side: content.SideBefore, Boundary: content.BoundaryInsideIfInsertedBefore},
		End:       content.AnchorPoint{RunID: order[3], BirthOrdinal: 9, Side: content.SideAfter, Boundary: content.BoundaryInsideIfInsertedAfter},
		BodyBlock: bodyBlock,
	}
	deleted := map[pdlfmt.UnitID]struct{}{order[2]: {}, order[3]: {}}
	ctx := content.OrphanContext{Author: 31415, QuotedText: "the deleted anchored passage, café résumé"}

	orphaned := content.OrphanOnDelete(ann, ctx, order, deleted)
	if !orphaned.Orphaned {
		t.Fatalf("annotation was not orphaned")
	}

	// "Save": encode the orphaned annotation and append it via the
	// append-only ledger placement API.
	payload := content.EncodeOrphanedAnnotation(nil, orphaned)
	prior := make([]byte, ledger.PrefixLength)
	res, err := ledger.PlaceWithinBudget(prior, 0, ledger.EditDelta{
		NewSegments: []ledger.SegmentPayload{{Octets: payload}},
	}, uint64(len(payload)))
	if err != nil {
		t.Fatalf("place (save): %v", err)
	}

	// "Reload": read the appended segment back out of the image and decode.
	reloadedBytes := res.Image[ledger.PrefixLength:]
	got, _, err := content.DecodeOrphanedAnnotation(reloadedBytes)
	if err != nil {
		t.Fatalf("decode after reload: %v", err)
	}

	// Zero-diff on all four FR-030 orphan fields.
	if got.Orphan.Author != orphaned.Orphan.Author {
		t.Fatalf("orphan author drifted: got %d, want %d", got.Orphan.Author, orphaned.Orphan.Author)
	}
	if got.Orphan.QuotedText != orphaned.Orphan.QuotedText {
		t.Fatalf("orphan quoted text drifted: got %q, want %q", got.Orphan.QuotedText, orphaned.Orphan.QuotedText)
	}
	if !got.Orphan.Prev.Equal(orphaned.Orphan.Prev) {
		t.Fatalf("orphan prev drifted")
	}
	if !got.Orphan.Next.Equal(orphaned.Orphan.Next) {
		t.Fatalf("orphan next drifted")
	}

	// Zero-diff on both boundary-behaviour values.
	if got.Start.Boundary != orphaned.Start.Boundary {
		t.Fatalf("start boundary drifted: got %v, want %v", got.Start.Boundary, orphaned.Start.Boundary)
	}
	if got.End.Boundary != orphaned.End.Boundary {
		t.Fatalf("end boundary drifted: got %v, want %v", got.End.Boundary, orphaned.End.Boundary)
	}

	// Full structural equality of the decoded annotation with the saved one.
	if got.ID != orphaned.ID || got.Start != orphaned.Start || got.End != orphaned.End ||
		got.BodyBlock != orphaned.BodyBlock || got.Orphan != orphaned.Orphan || !got.Orphaned {
		t.Fatalf("reloaded annotation differs from the saved one:\n got  %+v\n want %+v", got, orphaned)
	}
}
