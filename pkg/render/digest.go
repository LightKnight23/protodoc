// PageDirectory input-digest binding (T-0243; FR-086). The PageDirectory is a
// DERIVED, non-normative artefact; it is bound to a digest over the COMPLETE
// set of inputs it was computed from -- page geometry, the referenced content-
// unit identities, and the font digests -- so a reader can tell whether it is
// current (FR-086) and refuse it when stale (the staleness-refusal path is
// T-0244). This follows the
// input-digest-binding pattern used for PresentationArtefact (M09) and the
// Frontmatter preview (M01): a domain-tagged SHA-256 over a canonical,
// length-prefixed serialization of every input in a fixed order.
//
// The digest is a reader-side staleness check, NOT a signed value: the
// PageDirectory has no container.abnf wire discriminant and plays no role in
// any Merkle tree (data-model.md 2.22, CQ-016). Its domain prefix is a
// render-local multi-octet ASCII tag that cannot collide with integrity.abnf's
// single-octet preimage tags.
package render

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"

	"Protodoc/pkg/pdlfmt"
)

// pageDirectoryDigestDomain is the render-local domain prefix for the
// PageDirectory input digest. It is multi-octet ASCII so it cannot equal any
// single-octet integrity.abnf preimage tag.
var pageDirectoryDigestDomain = []byte("protodoc/render/page-directory-input-v1")

// PageDirectoryInputs is the COMPLETE declared input set a PageDirectory was
// computed from (FR-086). Recomputing the digest over an unchanged input set
// yields an identical value; changing any one input changes the digest.
type PageDirectoryInputs struct {
	// Geometry is the page-geometry token that governed pagination.
	Geometry uint32
	// Units are the referenced content-unit identities, in the order they were
	// consumed (order is part of the input identity and is preserved, not
	// sorted).
	Units []pdlfmt.UnitID
	// FontDigests are the digests of the fonts the pagination depended on.
	// Their SET (not order) is the input, so they are sorted before hashing.
	FontDigests []pdlfmt.Digest256
}

// Digest computes the FR-086 input digest over the complete declared input set,
// domain-tagged and canonically length-prefixed in a fixed field order:
// geometry, then the ordered unit identities, then the sorted font digests.
func (in PageDirectoryInputs) Digest() pdlfmt.Digest256 {
	h := sha256.New()
	h.Write(pageDirectoryDigestDomain)

	var u32 [4]byte
	binary.BigEndian.PutUint32(u32[:], in.Geometry)
	h.Write(u32[:])

	// Ordered unit identities, length-prefixed.
	binary.BigEndian.PutUint32(u32[:], uint32(len(in.Units)))
	h.Write(u32[:])
	for _, id := range in.Units {
		h.Write(id[:])
	}

	// Font digests as a SET: sort byte-lexicographically, length-prefixed.
	fonts := make([]pdlfmt.Digest256, len(in.FontDigests))
	copy(fonts, in.FontDigests)
	sort.Slice(fonts, func(i, j int) bool {
		for k := range fonts[i] {
			if fonts[i][k] != fonts[j][k] {
				return fonts[i][k] < fonts[j][k]
			}
		}
		return false
	})
	binary.BigEndian.PutUint32(u32[:], uint32(len(fonts)))
	h.Write(u32[:])
	for _, fd := range fonts {
		h.Write(fd[:])
	}

	var out pdlfmt.Digest256
	copy(out[:], h.Sum(nil))
	return out
}
