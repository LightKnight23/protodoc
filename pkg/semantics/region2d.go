// 2D presentation region (FR-099; T-0290, T-0291). For content that cannot be
// expressed as pure linear flow, a Region2D carries a MANDATORY linearised-
// reading-order reference into ROOT_SEQUENCE and a MANDATORY text alternative
// (reusing the T-0284 EmbeddedObjectAlt shape). Mirrors document.abnf S5.2.
//
// Provisional per T-0267 (clarify-002.md, OPEN awaiting Eyvar).
package semantics

import (
	"errors"

	"Protodoc/pkg/pdlfmt"
)

// Region2D is a 2D-presentation region record.
type Region2D struct {
	RegionID pdlfmt.UnitID
	// LinearRef is the ROOT_SEQUENCE unit-id at which this region is read.
	LinearRef pdlfmt.UnitID
	// Alt is the region's text alternative (decorative flag + text), reusing
	// the embedded-object alt-text shape (T-0284).
	Alt EmbeddedObjectAlt
}

// EncodeRegion2D appends the region's linear-ref (32 octets) followed by its
// alt-text encoding.
func EncodeRegion2D(dst []byte, r Region2D) []byte {
	dst = append(dst, r.LinearRef[:]...)
	dst = EncodeAlt(dst, r.Alt)
	return dst
}

// DecodeRegion2D reads a region's linear-ref and alt text from buf, returning
// the value and octets consumed. The region id is not part of this field
// bundle (it is the enclosing record's tag=1) and is left zero.
func DecodeRegion2D(buf []byte) (Region2D, int, error) {
	const idLen = 16
	if len(buf) < idLen {
		return Region2D{}, 0, errors.New("semantics: short region-2d buffer")
	}
	var r Region2D
	copy(r.LinearRef[:], buf[:idLen])
	alt, c, err := DecodeAlt(buf[idLen:])
	if err != nil {
		return Region2D{}, 0, err
	}
	r.Alt = alt
	return r, idLen + c, nil
}
