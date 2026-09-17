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

// DocumentOrder is the document's run identities in traversal/storage order,
// the deterministic sequence orphan resolution walks to find the surviving
// units nearest a deleted span. It is derived from the authoritative content
// (the run order), so it is identical for a fixed document state.
type DocumentOrder []pdlfmt.UnitID

// ResolveOrphan returns the nearest surviving unit before and after the
// orphaned annotation a, given the document's run order and the set of
// deleted run identities. Resolution is deterministic for a fixed document
// state: it walks order once and returns the same (prev, next) pair on every
// call. prev is the nearest surviving unit strictly before the annotation's
// start anchor position in order; next is the nearest surviving unit
// strictly after the end anchor position. Either is the zero unit-id when the
// orphaned span sat at the document start (no surviving predecessor) or end
// (no surviving successor). ResolveOrphan does not mutate a.
//
// Positions are located by the annotation's own anchor run identities within
// order (content identity, not a counted position).
func ResolveOrphan(a Annotation, order DocumentOrder, deleted map[pdlfmt.UnitID]struct{}) (prev, next pdlfmt.UnitID) {
	startIdx, endIdx := -1, -1
	for i, id := range order {
		if id.Equal(a.Start.RunID) {
			startIdx = i
		}
		if id.Equal(a.End.RunID) {
			endIdx = i
		}
	}
	if startIdx >= 0 {
		for i := startIdx - 1; i >= 0; i-- {
			if _, gone := deleted[order[i]]; !gone {
				prev = order[i]
				break
			}
		}
	}
	if endIdx >= 0 {
		for i := endIdx + 1; i < len(order); i++ {
			if _, gone := deleted[order[i]]; !gone {
				next = order[i]
				break
			}
		}
	}
	return prev, next
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
