// Cross-reference presentation function (FR-084; T-0288). A CROSS_REFERENCE
// declares HOW it renders — target page number, citation label, target
// numbering label, or target title text — and the rendered reference text is
// computed at materialisation as a COMPUTED_INLINE (T-0272), never a persisted
// frozen literal. This replaces any reliance on authored frozen reference text.
//
// Provisional per T-0267 (clarify-002.md, OPEN awaiting Eyvar).
package semantics

import (
	"errors"

	"Protodoc/pkg/pdlfmt"
)

// PresentationFn is the closed set of cross-reference presentation functions.
type PresentationFn uint8

const (
	// PresentTargetPage: render the target's page number.
	PresentTargetPage PresentationFn = 0x00
	// PresentCitationLabel: render the target's citation label.
	PresentCitationLabel PresentationFn = 0x01
	// PresentNumberingLabel: render the target's numbering label.
	PresentNumberingLabel PresentationFn = 0x02
	// PresentTitleText: render the target's title text.
	PresentTitleText PresentationFn = 0x03
)

// ErrPresentationFnOutOfSet rejects reserved presentation-function values.
var ErrPresentationFnOutOfSet = errors.New("semantics: xref presentation-fn outside closed value set {0..3}")

// ValidFn reports whether f is an assigned presentation function.
func (f PresentationFn) ValidFn() bool { return f <= PresentTitleText }

// CrossReference is the subset of a CROSS_REFERENCE the presentation model
// needs: its identity, its resolved target, and its presentation function.
type CrossReference struct {
	XrefID         pdlfmt.UnitID
	Target         pdlfmt.UnitID
	PresentationFn PresentationFn
}

// EncodePresentationFn appends the presentation-fn octet after validating it.
func EncodePresentationFn(dst []byte, fn PresentationFn) ([]byte, error) {
	if !fn.ValidFn() {
		return nil, ErrPresentationFnOutOfSet
	}
	return append(dst, byte(fn)), nil
}

// DecodePresentationFn reads the presentation-fn octet, rejecting reserved
// values, returning the value and octets consumed (1).
func DecodePresentationFn(buf []byte) (PresentationFn, int, error) {
	if len(buf) < 1 {
		return 0, 0, errors.New("semantics: short presentation-fn buffer")
	}
	fn := PresentationFn(buf[0])
	if !fn.ValidFn() {
		return 0, 0, ErrPresentationFnOutOfSet
	}
	return fn, 1, nil
}

// RenderReference materialises the reference's display text as a COMPUTED_INLINE
// given the resolved-target values it may present (page number, citation label,
// numbering label, title). The chosen value is a deterministic function of the
// presentation function and the resolved target — never authored frozen text.
func RenderReference(xr CrossReference, targetPage int, citation, numbering, title string) ComputedInline {
	var v string
	switch xr.PresentationFn {
	case PresentCitationLabel:
		v = citation
	case PresentNumberingLabel:
		v = numbering
	case PresentTitleText:
		v = title
	default: // PresentTargetPage
		v = pageString(targetPage)
	}
	return ComputedInline{Value: v, ValueDirection: DirectionLTR}
}

// pageString renders a page number without depending on strconv in the hot
// comment path (kept explicit for determinism).
func pageString(n int) string {
	if n < 0 {
		n = 0
	}
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
