package content

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFR_025_AddressingIsContentIdentityOnly is T-0075's named test. It
// statically asserts that text is addressed only by content identity
// (run_id, base_ordinal), never by a counted absolute text-unit position:
//
//   - Anchor is the content package's addressing type, carrying exactly
//     RunID and BaseOrdinal (both IDENTITY-COMPONENT).
//   - No exported struct field in the content package (nor, once they
//     exist, in the annotation/xref/table document-model consumers) is an
//     integer-typed ABSOLUTE-POSITION field: a field whose name denotes a
//     counted position/index/offset measured from a sequence start. The
//     only integer identity field permitted is base_ordinal, which is a
//     run-internal IDENTITY-COMPONENT, not a document-wide count.
func TestFR_025_AddressingIsContentIdentityOnly(t *testing.T) {
	// Anchor must exist and carry exactly the identity pair (compile-time
	// references to both fields).
	var a Anchor
	_ = a.RunID
	_ = a.BaseOrdinal

	// Directories to scan: the content package now, plus the document-model
	// consumer packages once they land (skipped silently if absent).
	dirs := []string{
		".",
		"../annotation",
		"../xref",
		"../table",
	}

	// Field names that denote a counted absolute position (forbidden as an
	// integer type). base_ordinal is explicitly allowed: it is a run-scoped
	// identity component, not an absolute count.
	forbiddenSubstrings := []string{
		"absoluteposition", "absoluteoffset", "charindex", "characterindex",
		"documentoffset", "docoffset", "globaloffset", "textposition",
		"absindex", "linearposition", "countedposition",
	}
	allowedIntFields := map[string]bool{
		"baseordinal": true, // IDENTITY-COMPONENT, run-scoped
	}

	intTypes := map[string]bool{
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"uintptr": true, "byte": true, "rune": true,
	}

	fset := token.NewFileSet()
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // consumer package not created yet
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
				continue
			}
			path := filepath.Join(dir, e.Name())
			f, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				t.Fatalf("parsing %s: %v", path, err)
			}
			ast.Inspect(f, func(n ast.Node) bool {
				ts, ok := n.(*ast.TypeSpec)
				if !ok || !ts.Name.IsExported() {
					return true
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					return true
				}
				for _, field := range st.Fields.List {
					ident, ok := field.Type.(*ast.Ident)
					if !ok || !intTypes[ident.Name] {
						continue
					}
					for _, name := range field.Names {
						if !name.IsExported() {
							continue
						}
						lower := strings.ToLower(name.Name)
						if allowedIntFields[lower] {
							continue
						}
						for _, bad := range forbiddenSubstrings {
							if strings.Contains(lower, bad) {
								t.Errorf("%s: exported struct %s has integer field %s (%s), an absolute counted position; text must be addressed by content identity only (FR-025)", path, ts.Name.Name, name.Name, ident.Name)
							}
						}
					}
				}
				return true
			})
		}
	}
}

// TestFR_025_AnchorResolvesByIdentity confirms an Anchor addresses content by
// identity: it equals another Anchor iff both run_id and base_ordinal match,
// and AnchorOf builds an anchor from a run-internal index without persisting
// that index.
func TestFR_025_AnchorResolvesByIdentity(t *testing.T) {
	rid, _ := testMintID()
	other, _ := testMintID()
	r := Run{RunID: rid, BaseOrdinal: 100, Text: "hello"}

	a, ok := AnchorOf(r, 2)
	if !ok {
		t.Fatalf("AnchorOf in range failed")
	}
	if a.RunID != rid || a.BaseOrdinal != 102 {
		t.Fatalf("AnchorOf(r,2) = %+v, want {rid, 102}", a)
	}
	if !a.Equal(Anchor{RunID: rid, BaseOrdinal: 102}) {
		t.Fatalf("Anchor.Equal false for identical anchors")
	}
	if a.Equal(Anchor{RunID: other, BaseOrdinal: 102}) {
		t.Fatalf("anchors with different run_id compared equal")
	}
	if a.Equal(Anchor{RunID: rid, BaseOrdinal: 103}) {
		t.Fatalf("anchors with different base_ordinal compared equal")
	}
	if _, ok := AnchorOf(r, 99); ok {
		t.Fatalf("AnchorOf out of range unexpectedly ok")
	}
}
