// Embedded-object text alternative (FR-040; T-0284, T-0285). A non-decorative
// embedded object (a content-bearing RASTER_IMAGE) carries a meaningful text
// alternative; a decorative one carries an empty alternative and is skipped by
// assistive technology. Mirrors document.abnf S7's ri-decorative / ri-alttext.
//
// Provisional per T-0267 (clarify-002.md, OPEN awaiting Eyvar).
package semantics

import (
	"errors"

	"Protodoc/pkg/pdlfmt"
)

// EmbeddedObjectAlt is the accessibility text alternative of an embedded
// object.
type EmbeddedObjectAlt struct {
	// Decorative is true for a purely decorative object (empty alt text).
	Decorative bool
	// AltText is the text alternative; non-empty for a non-decorative object,
	// empty for a decorative one.
	AltText string
}

// ErrShortAltBuffer is returned when a decode buffer is truncated.
var ErrShortAltBuffer = errors.New("semantics: short embedded-object alt-text buffer")

// EncodeAlt appends the decorative flag and the varint-length-prefixed alt
// text to dst.
func EncodeAlt(dst []byte, a EmbeddedObjectAlt) []byte {
	if a.Decorative {
		dst = append(dst, 0x01)
	} else {
		dst = append(dst, 0x00)
	}
	dst = pdlfmt.AppendVarint(dst, uint64(len(a.AltText)))
	dst = append(dst, a.AltText...)
	return dst
}

// DecodeAlt reads the decorative flag and the alt text from buf, returning the
// value and octets consumed.
func DecodeAlt(buf []byte) (EmbeddedObjectAlt, int, error) {
	if len(buf) < 1 {
		return EmbeddedObjectAlt{}, 0, ErrShortAltBuffer
	}
	var a EmbeddedObjectAlt
	a.Decorative = buf[0] == 0x01
	if buf[0] > 0x01 {
		return EmbeddedObjectAlt{}, 0, errors.New("semantics: reserved decorative flag value")
	}
	n, c, err := pdlfmt.DecodeVarint(buf[1:])
	if err != nil {
		return EmbeddedObjectAlt{}, 0, err
	}
	pos := 1 + c
	if pos+int(n) > len(buf) {
		return EmbeddedObjectAlt{}, 0, ErrShortAltBuffer
	}
	a.AltText = string(buf[pos : pos+int(n)])
	return a, pos + int(n), nil
}
