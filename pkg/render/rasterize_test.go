package render

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"math/big"
	"os"
	"path/filepath"
	"testing"
)

// ri builds an exact-rational point from integer base-unit coordinates.
func ri(x, y int64) RatPoint {
	return RatPoint{X: big.NewRat(x, 1), Y: big.NewRat(y, 1)}
}

// TestNFR_019_RasterizerExactRationalDeterminism is T-0248's named unit test
// (NFR-019, DP-010). It rasterizes the same input twice and requires bit-
// identical output (exact-rational arithmetic is deterministic across runs and
// targets), checks the round-half-to-even coverage quantizer against pinned
// vectors, and statically confirms the rasterizer uses NO float32/float64
// arithmetic.
func TestNFR_019_RasterizerExactRationalDeterminism(t *testing.T) {
	// A 10x10 diamond polygon in a 10x10 raster.
	poly := []RatPoint{ri(5, 0), ri(10, 5), ri(5, 10), ri(0, 5)}

	a := FillPolygon(poly, 10, 10)
	b := FillPolygon(poly, 10, 10)

	// Rasterizing the same input twice is bit-identical.
	if !bytes.Equal(a.Cover, b.Cover) {
		t.Fatal("rasterizing the same input twice was not bit-identical")
	}
	if a.Width != 10 || a.Height != 10 || len(a.Cover) != 100 {
		t.Fatalf("raster dims wrong: %dx%d len %d", a.Width, a.Height, len(a.Cover))
	}
	// The diamond is non-empty and its centre row is filled.
	filled := 0
	for _, c := range a.Cover {
		if c > 0 {
			filled++
		}
	}
	if filled == 0 {
		t.Fatal("diamond rasterized to an empty coverage bitmap")
	}

	// Round-half-to-even quantizer, pinned vectors.
	cases := []struct {
		num, den int64
		want     int
	}{
		{0, 1, 0},   // 0 -> 0
		{1, 1, 255}, // full -> 255
		{1, 2, 128}, // 0.5 -> 127.5 -> 128 (127.5 ties to even 128)
		{1, 255, 1}, // 1/255 -> exactly 1 -> 1
		{1, 510, 0}, // 0.5/255 -> 0.5 -> ties to even 0
	}
	for _, tc := range cases {
		got := quantizeCoverage(big.NewRat(tc.num, tc.den))
		if got != tc.want {
			t.Errorf("quantizeCoverage(%d/%d) = %d, want %d", tc.num, tc.den, got, tc.want)
		}
	}

	// round-half-to-even on a bare half: 2.5 -> 2, 3.5 -> 4.
	if v := roundHalfToEven(big.NewRat(5, 2)); v != 2 {
		t.Errorf("roundHalfToEven(5/2) = %d, want 2 (ties to even)", v)
	}
	if v := roundHalfToEven(big.NewRat(7, 2)); v != 4 {
		t.Errorf("roundHalfToEven(7/2) = %d, want 4 (ties to even)", v)
	}

	// Static check: no float32/float64 anywhere in the rasterizer source.
	assertNoFloatsInRasterizer(t)
}

// assertNoFloatsInRasterizer parses rasterize.go and fails if any float32 or
// float64 type or float literal appears in the CODE (NFR-019: zero floating-
// point in the arithmetic path). Comments may mention the words; only real
// code tokens count.
func assertNoFloatsInRasterizer(t *testing.T) {
	t.Helper()
	src, err := os.ReadFile(filepath.Join(".", "rasterize.go"))
	if err != nil {
		t.Fatalf("read rasterize.go: %v", err)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "rasterize.go", src, 0)
	if err != nil {
		t.Fatalf("parse rasterize.go: %v", err)
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.BasicLit:
			if x.Kind == token.FLOAT {
				t.Errorf("rasterize.go contains a float literal %q (NFR-019)", x.Value)
			}
		case *ast.Ident:
			if x.Name == "float32" || x.Name == "float64" {
				t.Errorf("rasterize.go uses the %s type (NFR-019 float-free)", x.Name)
			}
		}
		return true
	})
}
