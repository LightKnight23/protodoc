package pdlfmt

import "fmt"

// GeometricValue is a persisted geometric measurement (CON-012): an
// integer multiple of one base unit, 1/914400 inch. That base unit divides
// the inch, the point (1/72 inch), the millimetre and the 1/96-inch
// reference pixel exactly, so every common unit converts to it by exact
// integer multiplication, never a rounded approximation.
//
// GeometricValue is backed by int64 and carries no floating-point
// representation anywhere on the type -- no float32/float64 field, method
// or conversion is defined on it (CON-012's own verification clause, audit
// A-GEOM; also covered by the general CON-005 audit in pkg/container).
type GeometricValue int64

// UnitsPerInch is the number of GeometricValue base units in one inch
// (CON-012: 1/914400 inch).
const UnitsPerInch GeometricValue = 914400

// AppendGeometricValue appends v's PDL-VARINT encoding to dst (contracts/
// container.abnf S4 fm-page-dimensions NORMATIVE comment: field-value is a
// varint per dimension). A document dimension is never negative;
// AppendGeometricValue rejects a negative v rather than silently
// reinterpreting its two's-complement bit pattern as a large unsigned
// varint.
func AppendGeometricValue(dst []byte, v GeometricValue) ([]byte, error) {
	if v < 0 {
		return nil, fmt.Errorf("pdlfmt: geometric-value %d is negative", int64(v))
	}
	return AppendVarint(dst, uint64(v)), nil
}

// DecodeGeometricValue decodes a PDL-VARINT-encoded GeometricValue from the
// start of src, returning the value and octets consumed.
func DecodeGeometricValue(src []byte) (GeometricValue, int, error) {
	v, n, err := DecodeVarint(src)
	if err != nil {
		return 0, 0, fmt.Errorf("pdlfmt: geometric-value: %w", err)
	}
	if v > 1<<63-1 {
		return 0, 0, fmt.Errorf("pdlfmt: geometric-value %d exceeds int64 range", v)
	}
	return GeometricValue(v), n, nil
}
