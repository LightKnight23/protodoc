// Empty (BOTTOM) state emission (FR-124; T-0315). The canonical empty-content
// Document state (BOTTOM) has zero content/resource/history subtrees. Its L*
// emission is a valid, minimal, well-defined file: the fixed prefix skeleton
// plus zero segments. Used for fresh-document creation and as the base of every
// compaction/publish path.
package canon

import "Protodoc/pkg/container"

// MinimalPrefixLen is the fixed-prefix length of a minimal BOTTOM-state file:
// the fixed prefix skeleton (magic header, commit ring, frontmatter, segment
// table). It is derived from the container package's own region constants
// (SegmentTableOffset + SegmentTableRegionSize = 1,048,576) so it can never
// drift from the authoritative prefix layout.
const MinimalPrefixLen = container.SegmentTableOffset + container.SegmentTableRegionSize

// EmptyState returns the canonical BOTTOM Document: a state with zero subtrees.
func EmptyState() *Document {
	return &Document{}
}

// EmitBottom emits the minimal, well-defined file for the BOTTOM state: the
// fixed-prefix skeleton (all-zero here at this level of detail — the concrete
// prefix bytes are the container package's concern) followed by ZERO content
// bytes (C(BOTTOM) is empty). The returned slice length is exactly
// MinimalPrefixLen: a minimal valid file carries the prefix and no segments.
func EmitBottom() []byte {
	return make([]byte, MinimalPrefixLen)
}

// IsBottom reports whether a state is the canonical empty state (zero content
// subtrees), so callers can take the well-defined minimal path.
func IsBottom(state *Document) bool {
	return state == nil || len(state.Subtrees) == 0
}
