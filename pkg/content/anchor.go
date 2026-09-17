// Content-identity addressing (T-0075, FR-025). Anchor is the single shared
// address type every consumer uses to point at a range or point in text by
// CONTENT IDENTITY -- the pair (run_id, base_ordinal) -- rather than by a
// counted absolute text-unit position. Formatting ranges, comments, links,
// cross-references, bookmarks and change records all address text through
// Anchor; no consumer persists a counted position, because a counted
// position is relocated by the first concurrent edit (the offset-based
// failure this format exists to avoid). base_ordinal is an
// IDENTITY-COMPONENT (comparable and look-up-able), never a distance
// measured from a sequence start.
package content

import "Protodoc/pkg/pdlfmt"

// Anchor addresses a single point in text by content identity: the scalar
// value with identity (RunID, BaseOrdinal). A range is expressed as a pair
// of Anchors (start, end). Anchor carries no absolute/counted position:
// RunID is the minting identity of the run and BaseOrdinal is the
// run-internal scalar-value ordinal (both IDENTITY-COMPONENT), so an Anchor
// resolves to the same content after arbitrary surrounding edits.
type Anchor struct {
	RunID       pdlfmt.UnitID
	BaseOrdinal uint32
}

// Equal reports whether two anchors address the same point (same run
// identity and same base_ordinal). Anchors are compared by identity only,
// never by positional ordering.
func (a Anchor) Equal(other Anchor) bool {
	return a.RunID.Equal(other.RunID) && a.BaseOrdinal == other.BaseOrdinal
}

// AnchorOf returns the Anchor addressing the scalar value at run-internal
// index i within run r, and whether i is in range. It is the bridge from a
// run-scoped scalar index to a durable content-identity address; the index
// is consumed here and never stored, so no counted position escapes.
func AnchorOf(r Run, i int) (Anchor, bool) {
	rid, ord, ok := r.CharIdentityAt(i)
	if !ok {
		return Anchor{}, false
	}
	return Anchor{RunID: rid, BaseOrdinal: ord}, true
}
