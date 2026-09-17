// Annotation orphan carriage (T-0078, FR-028; document.abnf S3, S3.2).
// Deleting every content unit an annotation anchors to does NOT delete the
// annotation: it converts to an orphan-carriage record instead, so the
// annotation and its captured context survive rather than being silently
// dropped (DP-004 orphan carriage, R3 DELETE-DOMINATES). The nearest-
// surviving-unit resolution (T-0079) and full field capture (T-0080) build
// on this task's orphaning transition.
package content

import "Protodoc/pkg/pdlfmt"

// OrphanRecord captures the context an annotation needs after the content it
// anchored to is gone (document.abnf S3.2). Fields beyond Orphaned itself
// are populated by later tasks (T-0079 resolves Prev/Next; T-0080 captures
// Author/QuotedText); this task establishes the record and the orphaned
// transition.
type OrphanRecord struct {
	Author     uint16
	QuotedText string
	Prev       pdlfmt.UnitID
	Next       pdlfmt.UnitID
}

// Annotation is a durable, anchor-based reference into text (data-model.md
// S2.9). It anchors a range via Start/End AnchorPoints and carries its own
// body. When the entire anchored span is deleted, Orphaned becomes true and
// Orphan carries the retained context, rather than the annotation being
// removed from the document.
type Annotation struct {
	ID    pdlfmt.UnitID
	Start AnchorPoint
	End   AnchorPoint
	// BodyBlock is the unit-id of the TextBlock carrying the annotation's
	// own text.
	BodyBlock pdlfmt.UnitID

	// Orphaned reports whether the anchored span has been fully deleted.
	Orphaned bool
	// Orphan carries the retained context; meaningful only when Orphaned.
	Orphan OrphanRecord
}

// DeleteRange records the deletion of the set of run identities in
// deletedRunIDs from a document containing the given annotations, and
// returns the updated annotations. An annotation whose anchored span is
// ENTIRELY within the deleted set is converted to an orphan (Orphaned=true)
// rather than dropped; an annotation retaining at least one surviving
// anchor endpoint is left live. No annotation is ever removed from the
// returned slice: orphaning preserves it (FR-028).
//
// An annotation's anchored span is "fully deleted" when both its Start and
// End anchors reference runs in deletedRunIDs (the span had collapsed onto
// deleted content). A live endpoint on a surviving run keeps the annotation
// live. The input slice is not mutated.
func DeleteRange(annotations []Annotation, deletedRunIDs map[pdlfmt.UnitID]struct{}) []Annotation {
	out := make([]Annotation, len(annotations))
	copy(out, annotations)
	for i := range out {
		a := &out[i]
		if a.Orphaned {
			continue // already orphaned; deletion does not resurrect or drop it
		}
		_, startGone := deletedRunIDs[a.Start.RunID]
		_, endGone := deletedRunIDs[a.End.RunID]
		if startGone && endGone {
			// The entire anchored span was deleted: orphan, never drop.
			a.Orphaned = true
		}
	}
	return out
}
