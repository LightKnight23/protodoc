// CROSS_REFERENCE base record (T-0369, FR-084/FR-085; document.abnf S5,
// data-model.md S2.26). A cross-reference stores the identity of its target
// plus a named presentation function (xref-kind), never literal frozen text
// (FR-084); its staleness is determinable from identity/presence alone,
// without computing layout (FR-085).
package content

import (
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// XrefKind is the closed 3-value cross-reference presentation function
// (document.abnf S5 xref-kind): the named function that produces the
// reference's displayed value from its target, so nothing is frozen as text.
type XrefKind uint8

const (
	// XrefInternalHyperlink renders as a link to the target (0x00).
	XrefInternalHyperlink XrefKind = 0x00
	// XrefCitation renders as a citation to the target (0x01).
	XrefCitation XrefKind = 0x01
	// XrefTableOfContentsEntry renders as a TOC entry for the target (0x02).
	XrefTableOfContentsEntry XrefKind = 0x02
)

// CrossReference is the decoded CROSS_REFERENCE record. Target is the
// identity of the referenced unit; Kind is the named presentation function;
// Anchor is the reference's own source position (boundary NORMATIVE-ignored).
type CrossReference struct {
	ID     pdlfmt.UnitID
	Target pdlfmt.UnitID
	Kind   XrefKind
	Anchor AnchorPoint
}

var (
	// ErrInvalidXrefKind is returned for an xref-kind octet outside {0..2}.
	ErrInvalidXrefKind = errors.New("content: xref-kind outside the closed set {0,1,2}")
	// ErrXrefTargetUnresolved is returned when a cross-reference target does
	// not resolve to exactly one present unit (FR-108).
	ErrXrefTargetUnresolved = errors.New("content: cross-reference target does not resolve to exactly one present unit")
	// ErrXrefTruncated is returned when the encoding ends prematurely.
	ErrXrefTruncated = errors.New("content: cross-reference encoding truncated")
)

// xrefFixedLen: id(16) + target(16) + kind(1) + anchor(22) = 55.
const xrefFixedLen = 16 + 16 + 1 + AnchorPointSize

// EncodeCrossReference appends a canonical fixed-layout encoding of x to dst.
func EncodeCrossReference(dst []byte, x CrossReference) []byte {
	dst = pdlfmt.AppendUnitID(dst, x.ID)
	dst = pdlfmt.AppendUnitID(dst, x.Target)
	dst = append(dst, byte(x.Kind))
	dst = EncodeAnchorPoint(dst, x.Anchor)
	return dst
}

// DecodeCrossReference decodes and validates a CrossReference from the
// leading xrefFixedLen octets of src, rejecting an xref-kind outside {0..2}
// and an invalid anchor.
func DecodeCrossReference(src []byte) (CrossReference, int, error) {
	if len(src) < xrefFixedLen {
		return CrossReference{}, 0, ErrXrefTruncated
	}
	var x CrossReference
	pos := 0
	id, _, err := pdlfmt.DecodeUnitID(src[pos : pos+16])
	if err != nil {
		return CrossReference{}, 0, fmt.Errorf("content: xref-id: %w", err)
	}
	x.ID = id
	pos += 16

	target, _, err := pdlfmt.DecodeUnitID(src[pos : pos+16])
	if err != nil {
		return CrossReference{}, 0, fmt.Errorf("content: xref-target: %w", err)
	}
	x.Target = target
	pos += 16

	kind := src[pos]
	if kind > byte(XrefTableOfContentsEntry) {
		return CrossReference{}, 0, fmt.Errorf("%w: got 0x%02x", ErrInvalidXrefKind, kind)
	}
	x.Kind = XrefKind(kind)
	pos++

	anchor, adv, err := DecodeAnchorPoint(src[pos:])
	if err != nil {
		return CrossReference{}, 0, fmt.Errorf("content: xref-anchor: %w", err)
	}
	x.Anchor = anchor
	pos += adv

	return x, pos, nil
}

// ResolveTarget checks that x's target resolves to exactly one unit present
// in the document (FR-108), returning ErrXrefTargetUnresolved naming the
// reference otherwise. present maps every present unit id to how many times
// it occurs; exactly one occurrence is required (a dangling target has zero,
// an ambiguous one has more than one).
func (x CrossReference) ResolveTarget(present map[pdlfmt.UnitID]int) error {
	if present[x.Target] != 1 {
		return fmt.Errorf("%w: xref %x target %x resolves to %d units", ErrXrefTargetUnresolved, x.ID, x.Target, present[x.Target])
	}
	return nil
}

// IsStale reports whether x is stale, determined WITHOUT computing layout
// (FR-085): a reference is stale iff its target is no longer present in the
// document (present count != 1). This is a pure identity/presence check over
// the content inventory; it reaches no layout entry point, so every cheap
// consumer can compute it. currentPresence maps present unit ids to their
// occurrence count.
func (x CrossReference) IsStale(currentPresence map[pdlfmt.UnitID]int) bool {
	return currentPresence[x.Target] != 1
}
