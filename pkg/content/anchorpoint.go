// Anchor point wire form (document.abnf S3.1 anchor-point), reused by
// Annotation, Note, CrossReference and ExtEnvelope. It carries zero
// persisted counted positions: an anchor is (run-id, birth-ordinal) content
// identity plus a side and a boundary behaviour. This file provides the
// fixed-layout encode/decode used by the T-0077 boundary-behaviour
// save/load round-trip and by later annotation tasks.
package content

import (
	"encoding/binary"
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// AnchorSide is the closed 2-value anchor-side enum (document.abnf S3.1
// anchor-side): before or after the indexed character.
type AnchorSide uint8

const (
	// SideBefore: the anchor sits before the indexed character (0x00).
	SideBefore AnchorSide = 0x00
	// SideAfter: the anchor sits after the indexed character (0x01).
	SideAfter AnchorSide = 0x01
)

// AnchorPoint is the decoded anchor-point structure. BirthOrdinal is the
// within-run offset at anchor-creation time (IDENTITY-COMPONENT, subject to
// the same split/merge preservation as base_ordinal), never a
// document-relative count.
type AnchorPoint struct {
	RunID        pdlfmt.UnitID
	BirthOrdinal uint32
	Side         AnchorSide
	Boundary     AnchorBoundary
}

// AnchorPointSize is anchor-point's fixed encoded width: unit-id(16) +
// birth-ordinal(4) + side(1) + boundary(1) = 22 octets.
const AnchorPointSize = 16 + 4 + 1 + 1

var (
	// ErrAnchorPointTruncated is returned when fewer than AnchorPointSize
	// octets are available to decode.
	ErrAnchorPointTruncated = errors.New("content: anchor-point truncated, need 22 octets")
	// ErrInvalidAnchorSide is returned for an anchor-side octet outside the
	// closed set {0x00, 0x01}.
	ErrInvalidAnchorSide = errors.New("content: anchor-side outside the closed set {0,1}")
)

// EncodeAnchorPoint appends a's canonical 22-octet encoding to dst. It does
// not validate a's Side/Boundary against the closed enums (a caller
// constructing an AnchorPoint directly supplies valid values); DecodeAnchorPoint
// validates on the way in, at the trust boundary.
func EncodeAnchorPoint(dst []byte, a AnchorPoint) []byte {
	dst = pdlfmt.AppendUnitID(dst, a.RunID)
	var ord [4]byte
	binary.BigEndian.PutUint32(ord[:], a.BirthOrdinal)
	dst = append(dst, ord[:]...)
	dst = append(dst, byte(a.Side), byte(a.Boundary))
	return dst
}

// DecodeAnchorPoint decodes and validates an anchor-point from the leading
// AnchorPointSize octets of src, rejecting an invalid side (not in {0,1}) or
// an invalid boundary (rule PD-BOUNDARY-001). It returns the decoded point
// and the number of octets consumed (always AnchorPointSize on success).
func DecodeAnchorPoint(src []byte) (AnchorPoint, int, error) {
	if len(src) < AnchorPointSize {
		return AnchorPoint{}, 0, ErrAnchorPointTruncated
	}
	var a AnchorPoint
	id, _, err := pdlfmt.DecodeUnitID(src[:16])
	if err != nil {
		return AnchorPoint{}, 0, fmt.Errorf("content: anchor-point run-id: %w", err)
	}
	a.RunID = id
	a.BirthOrdinal = binary.BigEndian.Uint32(src[16:20])

	side := src[20]
	if side != byte(SideBefore) && side != byte(SideAfter) {
		return AnchorPoint{}, 0, fmt.Errorf("%w: got 0x%02x", ErrInvalidAnchorSide, side)
	}
	a.Side = AnchorSide(side)

	boundary, err := DecodeAnchorBoundary(src[21])
	if err != nil {
		return AnchorPoint{}, 0, err
	}
	a.Boundary = boundary
	return a, AnchorPointSize, nil
}
