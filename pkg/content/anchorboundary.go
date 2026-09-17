// Annotation anchor boundary behaviour (T-0076, FR-026; document.abnf S3.1
// anchor-boundary). Each annotation endpoint declares exactly one of a
// CLOSED 4-value boundary behaviour, governing whether text inserted at the
// endpoint joins the annotated range. Decoding a 5th/undefined value is a
// structural rejection (rule PD-BOUNDARY-001), so the enum can never silently
// admit an unrecognised behaviour.
package content

import (
	"errors"
	"fmt"
)

// AnchorBoundary is the closed boundary-behaviour enum (document.abnf S3.1).
type AnchorBoundary uint8

const (
	// BoundaryInside: the endpoint stays inside the range (0x00).
	BoundaryInside AnchorBoundary = 0x00
	// BoundaryOutside: the endpoint stays outside the range (0x01).
	BoundaryOutside AnchorBoundary = 0x01
	// BoundaryInsideIfInsertedBefore: text inserted before this exact anchor
	// point joins the range (0x02).
	BoundaryInsideIfInsertedBefore AnchorBoundary = 0x02
	// BoundaryInsideIfInsertedAfter: symmetric, for insertion after (0x03).
	BoundaryInsideIfInsertedAfter AnchorBoundary = 0x03

	// boundaryMaxValid is the highest defined value; 0x04..0xFF are reserved.
	boundaryMaxValid = 0x03
)

// ErrInvalidAnchorBoundary is rule PD-BOUNDARY-001: an anchor-boundary octet
// outside the closed set {0x00..0x03} is rejected.
var ErrInvalidAnchorBoundary = errors.New("content: anchor-boundary outside the closed set {0..3} (rule PD-BOUNDARY-001)")

// valid reports whether b is one of the four defined boundary behaviours.
func (b AnchorBoundary) valid() bool { return b <= boundaryMaxValid }

// String renders the boundary behaviour for diagnostics.
func (b AnchorBoundary) String() string {
	switch b {
	case BoundaryInside:
		return "inside"
	case BoundaryOutside:
		return "outside"
	case BoundaryInsideIfInsertedBefore:
		return "inside-if-inserted-before"
	case BoundaryInsideIfInsertedAfter:
		return "inside-if-inserted-after"
	default:
		return fmt.Sprintf("invalid(0x%02x)", uint8(b))
	}
}

// DecodeAnchorBoundary decodes and validates one anchor-boundary octet,
// rejecting any value in the reserved range 0x04..0xFF with a
// PD-BOUNDARY-001 error (structural reject, no silent skip).
func DecodeAnchorBoundary(octet byte) (AnchorBoundary, error) {
	b := AnchorBoundary(octet)
	if !b.valid() {
		return 0, fmt.Errorf("%w: got 0x%02x", ErrInvalidAnchorBoundary, octet)
	}
	return b, nil
}
