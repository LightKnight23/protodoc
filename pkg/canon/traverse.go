// Package canon implements Protodoc canonicalization C(S) and the compaction
// operations built on it (NFR-002, NFR-004). This file provides the core
// depth-first traversal (T-0307): it walks a Document state's content subtrees
// in the SAME canonical order T_C establishes (integrity.abnf S2.2.1 —
// ascending unit-id when no ROOT_SEQUENCE keys the order), yielding the ordered
// sequence of PDL-TLV records constituting C(S). This walk is shared by
// Canonicalize, Compact, PartialCompact, Publish, Migrate and the projectors.
package canon

import (
	"bytes"
	"sort"

	"Protodoc/pkg/pdlfmt"
)

// ContentSubtree is one content-model record's canonical unit: its unit-id
// (which fixes its position under the canonical order) and its already-canonical
// PDL-TLV frame bytes.
type ContentSubtree struct {
	UnitID pdlfmt.UnitID
	Frame  []byte
}

// Document is a logical document state S: its content subtrees in ARBITRARY
// storage order. Canonicalization is a pure function of this logical state and
// is independent of the slice order here (storage-order independence, NFR-002).
type Document struct {
	Subtrees []ContentSubtree
}

// canonicalLess is T_C's per-subtree canonical order: ascending unsigned
// byte-lexicographic order of unit-id (integrity.abnf S2.2.1 interim rule).
func canonicalLess(a, b ContentSubtree) bool {
	return bytes.Compare(a.UnitID[:], b.UnitID[:]) < 0
}

// Traverse returns the document's content subtrees in canonical (T_C) order: a
// new slice sorted by ascending unit-id, never mutating the input. The BOTTOM
// (empty) state yields zero content records. The order is a pure function of
// the logical state, so two calls on the same state — and on any storage-order
// permutation of it — produce the identical ordered sequence.
func Traverse(state *Document) []ContentSubtree {
	if state == nil || len(state.Subtrees) == 0 {
		return nil
	}
	out := make([]ContentSubtree, len(state.Subtrees))
	copy(out, state.Subtrees)
	sort.SliceStable(out, func(i, j int) bool { return canonicalLess(out[i], out[j]) })
	return out
}
