// Glyph outline evaluation without grid-fitting or instruction execution
// (T-0249; NFR-023, CQ-006/CP-005). Reference rasterisation evaluates glyph
// outlines as OPAQUE outline data only: it never grid-fits (no hinting-driven
// point movement) and never executes any instruction stream a font may carry.
// This upholds the data-not-a-program rule and keeps coverage-per-glyph-per-
// size implementation-independent -- and there is deliberately no instruction
// interpreter anywhere in this package.
package render

import "math/big"

// GlyphSubset is an embedded font subset as the rasterizer sees it: the outline
// contours (the only thing evaluated) and an OPAQUE instruction stream that is
// carried but never interpreted. StripInstructions produces the same subset
// with the stream removed; both must render identically (NFR-023).
type GlyphSubset struct {
	// Contours are the glyph's closed outline contours in base units. This is
	// the sole input to outline evaluation.
	Contours [][]RatPoint
	// Instructions is the font's hinting/instruction stream. The reference
	// rasterizer NEVER reads this; it exists only so the T-0249 test can prove
	// its presence changes nothing.
	Instructions []byte
}

// StripInstructions returns a copy of the subset with the instruction stream
// removed. The contours are unchanged.
func (g GlyphSubset) StripInstructions() GlyphSubset {
	return GlyphSubset{Contours: g.Contours}
}

// EvaluateOutline rasterizes a glyph subset's outline into a coverage Raster of
// the given size WITHOUT grid-fitting and WITHOUT touching the instruction
// stream. It fills every contour with non-zero winding via the exact-rational
// FillPolygon. Because g.Instructions is never read and no point is moved by a
// hinting program, the result depends only on the outline geometry.
func EvaluateOutline(g GlyphSubset, width, height int) Raster {
	r := Raster{Width: width, Height: height, Cover: make([]byte, width*height)}
	for _, contour := range g.Contours {
		filled := FillPolygon(contour, width, height)
		for i := range r.Cover {
			total := int(r.Cover[i]) + int(filled.Cover[i])
			if total > 255 {
				total = 255
			}
			r.Cover[i] = byte(total)
		}
	}
	return r
}

// gridFitDisabled is a compile-time-visible statement of intent: the outline
// evaluator applies no grid-fitting. The sample grid is fixed at half-integer
// scanline centres (see rasterize.go) with no per-glyph adjustment. It is a
// value so the T-0249 test can assert the policy is present and true.
var gridFitDisabled = big.NewRat(0, 1) // zero grid adjustment, always
