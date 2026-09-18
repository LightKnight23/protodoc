package render

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// TestEX001_ShapingIsolationDoesNotBlockCoreRenderPath is T-0266's named
// integration test (EX-001 / CON-006 blast-radius containment). The core render
// path -- rasterizer, PLP-1/PNG decode, reflow, and fixed pagination -- builds
// and runs green independently of the glyph-shaping implementation: it produces
// identical results whether the Shaper is the pinned-oracle implementation or a
// stub, because those pieces never consult a Shaper at all. A static scan
// confirms the core-path source files do not reference the shaping symbols.
func TestEX001_ShapingIsolationDoesNotBlockCoreRenderPath(t *testing.T) {
	// Both a real and a stub Shaper exist and satisfy the interface.
	var _ Shaper = PinnedOracleShaper{}
	var _ Shaper = StubShaper{}

	// The stub returns no glyphs; the real one returns some.
	in := ShapingInput{Scalars: []rune("hi")}
	if len(StubShaper{}.ShapeRun(1, in)) != 0 {
		t.Error("stub shaper must return no glyphs")
	}
	if len(PinnedOracleShaper{}.ShapeRun(1, in)) == 0 {
		t.Error("pinned-oracle shaper must return glyphs")
	}

	// Core render path runs green regardless of shaping: rasterize, decode,
	// reflow, paginate all succeed with NO Shaper involved.

	// Rasterizer.
	if r := FillPolygon([]RatPoint{ri(2, 0), ri(8, 6), ri(0, 6)}, 8, 8); len(r.Cover) != 64 {
		t.Error("rasterizer must run without shaping")
	}
	// PLP-1 decode.
	if _, _, _, err := DecodePLP1(buildValidVector("gradient_8x8", 8, 8, 3)); err != nil {
		t.Errorf("PLP-1 decode must run without shaping: %v", err)
	}
	// Restricted-PNG decode.
	if _, err := DecodeRestrictedPNG(encodeRestrictedPNG(2, 2, 3, make([]byte, 12))); err != nil {
		t.Errorf("PNG decode must run without shaping: %v", err)
	}
	// Reflow.
	if len(Reflow(buildParagraph(10, 1000, 300), 5000)) == 0 {
		t.Error("reflow must run without shaping")
	}
	// Fixed pagination.
	fm := &container.Frontmatter{PageCount: 3, PageWidth: pdlfmt.GeometricValue(1000), PageHeight: pdlfmt.GeometricValue(2000)}
	if ProjectFixedPagination(fm, nil).PageCount != 3 {
		t.Error("fixed pagination must run without shaping")
	}

	// Static: the core-path source files never reference the shaping symbols.
	assertCorePathShapingIndependent(t)
}

// assertCorePathShapingIndependent parses the core-render-path source files and
// fails if any references Shape/Shaper/ShapingInput -- proving the shaping
// sub-slice is isolated behind its interface.
func assertCorePathShapingIndependent(t *testing.T) {
	t.Helper()
	corePath := []string{
		"rasterize.go", "plp1.go", "restrictedpng.go", "reflow.go",
		"reflowfit.go", "pagination.go", "pagedirectory.go",
	}
	shapingSymbols := map[string]bool{
		"Shape": true, "Shaper": true, "ShapingInput": true,
		"ShapedGlyph": true, "ShapingProfileID": true,
		"PinnedOracleShaper": true, "StubShaper": true,
	}
	fset := token.NewFileSet()
	for _, name := range corePath {
		src, err := os.ReadFile(filepath.Join(".", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		// A cheap textual guard too (catches references in any form).
		if bytes.Contains(src, []byte("Shaper")) {
			t.Errorf("%s references the shaping interface (core path must be shaping-independent)", name)
		}
		f, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && shapingSymbols[id.Name] {
				t.Errorf("%s references shaping symbol %q (core path must be shaping-independent)", name, id.Name)
			}
			return true
		})
	}
}
