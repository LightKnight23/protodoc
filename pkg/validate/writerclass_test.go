package validate

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// TestCON_017_SingleWriterConformanceClassAsserted is T-0126's named unit
// test (the A-CLASS audit's CI lint guard). CON-017 requires the format
// define EXACTLY ONE writer conformance class: any document produced by any
// conforming writer must satisfy the identical validation pipeline, with no
// separate lenient/strict/legacy/transitional writer-class flag or branch
// anywhere in the validate package. The requirement is satisfied by the
// ABSENCE of a second class, not by new logic, so this test is a guard that
// FAILS CI if anyone introduces a writer-class concept: it scans every
// non-test .go file in the validation package for (a) any identifier that
// names a writer class or class-tier, and (b) any conditional branch keyed on
// such a concept. Finding one means a second class was smuggled in.
func TestCON_017_SingleWriterConformanceClassAsserted(t *testing.T) {
	// Tokens that would signal a writer-class or class-tier concept. Matching
	// is case-insensitive on identifier text. "strict"/"lenient" are the
	// canonical two-tier split the markup incumbent shipped; the others are
	// the usual euphemisms for a retired-but-default class.
	forbiddenSubstrings := []string{
		"writerclass", "writerconformanceclass", "writer_class",
		"lenientwriter", "strictwriter", "legacywriter", "transitionalwriter",
		"migrationclass", "compatibilityclass", "classtier", "writertier",
	}
	// Bare tier words that, combined with "class"/"mode"/"writer" context in
	// the same identifier, indicate a tiered writer class. We match these only
	// as whole identifier fragments to avoid flagging unrelated words.
	tierWords := []string{"lenient", "strict", "transitional", "legacy"}

	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read validate package dir: %v", err)
	}
	scanned := 0
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		scanned++
		af, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		ast.Inspect(af, func(n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if !ok {
				return true
			}
			low := strings.ToLower(id.Name)
			for _, bad := range forbiddenSubstrings {
				if strings.Contains(low, bad) {
					t.Errorf("%s: identifier %q names a writer-class concept; CON-017 permits exactly one writer conformance class (no second/tiered class)", name, id.Name)
				}
			}
			// A tier word is only suspicious when combined with a class/writer
			// context in the same identifier (e.g. "strictClass", "writerLenient").
			for _, tw := range tierWords {
				if strings.Contains(low, tw) && (strings.Contains(low, "class") || strings.Contains(low, "writer") || strings.Contains(low, "mode")) {
					t.Errorf("%s: identifier %q pairs a class-tier word with a writer/class context; CON-017 forbids a second writer conformance class", name, id.Name)
				}
			}
			return true
		})
	}
	if scanned == 0 {
		t.Fatal("A-CLASS lint scanned no validate-package source files")
	}
}
