package pdlfmt

import (
	"reflect"
	"testing"
)

// TestCON_012_GeometricValueIsFixedPointInteger is T-0027's named test.
// Implements: CON-012.
func TestCON_012_GeometricValueIsFixedPointInteger(t *testing.T) {
	gvType := reflect.TypeOf(GeometricValue(0))
	// GeometricValue is a scalar named type over int64 (not a struct), so
	// there is only one kind to check: it must be Int64, which by
	// construction rules out Float32/Float64 -- there is no second field to
	// walk (CON-012: no floating-point representation anywhere on the type).
	if gvType.Kind() != reflect.Int64 {
		t.Fatalf("GeometricValue's underlying kind is %s, want int64 (CON-012: no floating-point representation)", gvType.Kind())
	}

	cases := []struct {
		name          string
		width, height GeometricValue
	}{
		{"US Letter (8in x 11in)", 8 * UnitsPerInch, 11 * UnitsPerInch},
		{"A4 (210mm x 297mm, 36000 units/mm)", 210 * 36000, 297 * 36000},
		{"zero", 0, 0},
		{"single base unit", 1, 1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf []byte
			buf, err := AppendGeometricValue(buf, c.width)
			if err != nil {
				t.Fatalf("AppendGeometricValue(width=%d): %v", c.width, err)
			}
			buf, err = AppendGeometricValue(buf, c.height)
			if err != nil {
				t.Fatalf("AppendGeometricValue(height=%d): %v", c.height, err)
			}

			gotW, n1, err := DecodeGeometricValue(buf)
			if err != nil {
				t.Fatalf("DecodeGeometricValue(width): %v", err)
			}
			gotH, n2, err := DecodeGeometricValue(buf[n1:])
			if err != nil {
				t.Fatalf("DecodeGeometricValue(height): %v", err)
			}
			if n1+n2 != len(buf) {
				t.Fatalf("consumed %d octets, encoded %d", n1+n2, len(buf))
			}
			if gotW != c.width || gotH != c.height {
				t.Errorf("round-trip = (%d,%d), want (%d,%d)", gotW, gotH, c.width, c.height)
			}
		})
	}

	if _, err := AppendGeometricValue(nil, -1); err == nil {
		t.Error("AppendGeometricValue(-1) succeeded, want an error: a document dimension is never negative")
	}
}
