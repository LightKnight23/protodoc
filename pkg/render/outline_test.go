package render

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNFR_023_NoGridFittingNoInstructionExecution is T-0249's named unit test
// (NFR-023). A crafted font subset carrying an instruction stream renders
// identically to the same subset with the stream stripped (the stream is never
// executed), and a static scan confirms there is zero instruction-interpreter
// code in the render package.
func TestNFR_023_NoGridFittingNoInstructionExecution(t *testing.T) {
	// A glyph with one triangular contour, plus a bogus "instruction stream".
	withInstr := GlyphSubset{
		Contours: [][]RatPoint{{ri(2, 0), ri(10, 8), ri(0, 8)}},
		Instructions: []byte{
			// arbitrary hinting bytecode; a conforming rasterizer must ignore it.
			0xB0, 0x00, 0x2C, 0x01, 0xFF, 0x4B, 0x53, 0x23,
		},
	}
	stripped := withInstr.StripInstructions()
	if len(stripped.Instructions) != 0 {
		t.Fatal("StripInstructions must remove the instruction stream")
	}

	a := EvaluateOutline(withInstr, 10, 10)
	b := EvaluateOutline(stripped, 10, 10)

	// Identical rasters: the instruction stream changed nothing (NFR-023).
	if !bytes.Equal(a.Cover, b.Cover) {
		t.Error("subset with instruction stream must render identically to the stripped subset")
	}

	// A DIFFERENT instruction stream also changes nothing (not just absence).
	other := withInstr
	other.Instructions = bytes.Repeat([]byte{0x99}, 32)
	c := EvaluateOutline(other, 10, 10)
	if !bytes.Equal(a.Cover, c.Cover) {
		t.Error("changing the instruction stream must not change the raster")
	}

	// No grid-fitting adjustment is applied.
	if gridFitDisabled.Sign() != 0 {
		t.Error("grid-fitting must be disabled (zero adjustment)")
	}

	// Static analysis: zero instruction-interpreter code paths in the package.
	assertNoInterpreter(t)
}

// assertNoInterpreter scans every non-test .go file in the render package for
// signs of a font-instruction interpreter (an execute/interpret function over
// the Instructions field). NFR-023 requires none to exist.
func assertNoInterpreter(t *testing.T) {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(".", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		f, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			ln := strings.ToLower(fn.Name.Name)
			// Ban interpreter-shaped names that would execute a stream.
			for _, bad := range []string{"interpret", "execinstruction", "executeinstruction", "runinstruction", "vmstep", "hintprogram"} {
				if strings.Contains(ln, bad) {
					t.Errorf("%s: function %q looks like an instruction interpreter (NFR-023 forbids one)", name, fn.Name.Name)
				}
			}
			return true
		})
	}
}
