// Cross-reference staleness without layout (FR-085; T-0289). A CROSS_REFERENCE
// binds a digest to its target unit's state at materialisation time, the same
// input-digest-binding pattern PresentationArtefact / Frontmatter-preview /
// PageDirectory / UnitIndex use. A reader detects a STALE reference by
// comparing that bound digest against the target's CURRENT state digest alone —
// no layout is computed.
//
// Provisional per T-0267 (clarify-002.md, OPEN awaiting Eyvar).
package semantics

import "Protodoc/pkg/pdlfmt"

// RuleXrefStaleness is the validator rule id family for xref staleness (FR-085;
// tracked under the reference-resolution requirement, not a PD-rule of its own).
const xrefStalenessTask = "T-0289"

// BoundCrossReference is a cross-reference carrying the digest it bound to its
// target's state at materialisation.
type BoundCrossReference struct {
	XrefID      pdlfmt.UnitID
	Target      pdlfmt.UnitID
	BoundDigest pdlfmt.Digest256
}

// XrefStalenessFinding reports a stale cross-reference (bound digest != the
// target's current state digest), naming the reference and its target. It
// carries no layout — staleness is decided from the digest mismatch alone.
type XrefStalenessFinding struct {
	Task          string
	XrefID        pdlfmt.UnitID
	Target        pdlfmt.UnitID
	BoundDigest   pdlfmt.Digest256
	CurrentDigest pdlfmt.Digest256
}

// CheckXrefStaleness reports every cross-reference whose bound target digest no
// longer matches the target's current state digest (looked up in currentDigest
// by target id). A reference whose target is absent from currentDigest is also
// stale (its target's state cannot be confirmed). No layout is computed: the
// decision is purely the digest comparison. An empty result means every
// reference is current.
func CheckXrefStaleness(refs []BoundCrossReference, currentDigest map[pdlfmt.UnitID]pdlfmt.Digest256) []XrefStalenessFinding {
	var findings []XrefStalenessFinding
	for _, r := range refs {
		cur, ok := currentDigest[r.Target]
		if !ok || cur != r.BoundDigest {
			findings = append(findings, XrefStalenessFinding{
				Task: xrefStalenessTask, XrefID: r.XrefID, Target: r.Target,
				BoundDigest: r.BoundDigest, CurrentDigest: cur,
			})
		}
	}
	return findings
}
