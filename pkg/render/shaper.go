// Shaping isolation (T-0266; EX-001 / CP-009 blast-radius containment). The
// CP-009 shaping-oracle exception (standing constitutional text, recorded by
// T-0253) gates ONLY the glyph-shaping sub-slice, not the rasterizer, the
// PLP-1/PNG decoders, reflow, or fixed pagination. To keep the exception's
// blast radius contained, glyph shaping sits behind the Shaper interface that
// the rest of the pipeline does not depend on: the core render path (rasterize,
// decode, reflow, paginate) builds and runs with the Shaper stubbed or
// unimplemented.
package render

// Shaper is the narrow interface the render pipeline uses for glyph shaping.
// The concrete pinned-oracle implementation (Shape, T-0252) satisfies it, but
// the rasterizer, codecs, reflow, and pagination never reference this interface
// -- so the shaping sub-slice can be stubbed without affecting them.
type Shaper interface {
	ShapeRun(profile ShapingProfileID, in ShapingInput) []ShapedGlyph
}

// PinnedOracleShaper is the production Shaper backed by the pinned-oracle Shape
// function (T-0252).
type PinnedOracleShaper struct{}

// ShapeRun delegates to the pinned-oracle Shape.
func (PinnedOracleShaper) ShapeRun(profile ShapingProfileID, in ShapingInput) []ShapedGlyph {
	return Shape(profile, in)
}

// StubShaper is a Shaper that returns no glyphs. It exists to demonstrate that
// the core render path (rasterize / decode / reflow / paginate) is independent
// of any real shaping implementation: swapping in the stub does not change the
// rasterizer, codec, reflow, or pagination behaviour at all.
type StubShaper struct{}

// ShapeRun returns nil -- an unimplemented shaping sub-slice.
func (StubShaper) ShapeRun(profile ShapingProfileID, in ShapingInput) []ShapedGlyph {
	return nil
}
