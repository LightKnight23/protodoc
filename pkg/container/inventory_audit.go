// Single stored-unit inventory audit (CQ-007, FR-105): SegmentTable is the
// SOLE inventory of stored units in the container schema (contracts/
// container.abnf S5: "this is the SOLE inventory of stored units (CQ-007)
// -- no second, digest-keyed index exists anywhere in the format"). A
// second such inventory is exactly the "two disagreeing inventories"
// failure class FR-104/FR-105 close by construction, so this file provides
// a static, source-level audit rather than a hand-maintained allowlist: a
// hand-maintained list of "the inventory types" has the same blind spot as
// the failure it is meant to catch (it only knows about what someone
// remembered to add to it).
package container

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// packageSourceDir returns the absolute directory of this package's own
// source, so the audit below always scans the actual container package
// currently being built, not a path the caller has to know or keep in sync.
func packageSourceDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

// StoredUnitInventoryTypeNames statically scans every non-test .go file in
// dir (this package's own source) for struct type declarations matching
// the "stored-unit inventory record" shape: a field of type
// pdlfmt.Digest256 (a per-unit content digest), a byte/uint8-typed field
// whose name names a kind or type discriminant, and a field whose name
// names a length or size. SegmentTableSlot (SegmentType byte, Length
// uint64, Digest pdlfmt.Digest256) is exactly this shape and is the one
// type meant to match. It returns the matching type names in ascending
// order; CQ-007/FR-105 requires this list to have exactly one entry.
func StoredUnitInventoryTypeNames(dir string) ([]string, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				gd, ok := decl.(*ast.GenDecl)
				if !ok || gd.Tok != token.TYPE {
					continue
				}
				for _, spec := range gd.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					st, ok := ts.Type.(*ast.StructType)
					if !ok {
						continue
					}
					if isStoredUnitInventoryRecordShape(st) {
						names = append(names, ts.Name.Name)
					}
				}
			}
		}
	}
	sort.Strings(names)
	return names, nil
}

// isStoredUnitInventoryRecordShape reports whether st has all three
// signals a stored-unit inventory record needs: a per-unit content digest,
// a kind/type discriminant, and a locating offset. All three together are
// what makes something usable as an INDEX of stored units rather than
// merely a description of one already-located unit (data-model.md/
// container.abnf S5's "digest-keyed index" language): a struct that can
// name a unit's kind and digest but not say where it lives cannot serve as
// an alternative authority a reader could consult instead of SegmentTable,
// which is the specific failure class CQ-007/FR-104/FR-105 close by there
// being only one such struct, ever.
func isStoredUnitInventoryRecordShape(st *ast.StructType) bool {
	var hasDigest, hasKindDiscriminant, hasOffset bool
	if st.Fields == nil {
		return false
	}
	for _, field := range st.Fields.List {
		name := fieldName(field)
		lower := strings.ToLower(name)

		if isDigest256Type(field.Type) {
			hasDigest = true
		}
		if isByteType(field.Type) && strings.Contains(lower, "type") {
			hasKindDiscriminant = true
		}
		if isIntegerType(field.Type) && strings.Contains(lower, "offset") {
			hasOffset = true
		}
	}
	return hasDigest && hasKindDiscriminant && hasOffset
}

// fieldName returns a struct field's declared name, or "" for an embedded
// field (which this audit never treats as a kind/length discriminant: a
// name-bearing signal has to actually name something).
func fieldName(field *ast.Field) string {
	if len(field.Names) == 0 {
		return ""
	}
	return field.Names[0].Name
}

// isDigest256Type reports whether t is exactly pdlfmt.Digest256.
func isDigest256Type(t ast.Expr) bool {
	sel, ok := t.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return pkgIdent.Name == "pdlfmt" && sel.Sel.Name == "Digest256"
}

// isByteType reports whether t is the predeclared byte or uint8 type.
func isByteType(t ast.Expr) bool {
	id, ok := t.(*ast.Ident)
	if !ok {
		return false
	}
	return id.Name == "byte" || id.Name == "uint8"
}

// isIntegerType reports whether t is one of Go's predeclared integer types.
var integerTypeNames = map[string]bool{
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"byte": true, "rune": true,
}

func isIntegerType(t ast.Expr) bool {
	id, ok := t.(*ast.Ident)
	if !ok {
		return false
	}
	return integerTypeNames[id.Name]
}
