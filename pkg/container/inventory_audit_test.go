package container

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// TestFR_105_ExactlyOneStoredUnitInventoryType is T-0020's named test.
// Implements: FR-105.
//
// This is a schema-level, static test: it parses the container package's
// own source (not a hand-maintained list of "the inventory types") and
// asserts exactly one declared struct type has the shape of a stored-unit
// inventory record (a per-unit digest, a kind discriminant and a locating
// offset). CQ-007/FR-105 require SegmentTableSlot to be the sole such
// type; adding a second matching type changes this count and fails CI
// without anyone having to remember to update an allowlist.
func TestFR_105_ExactlyOneStoredUnitInventoryType(t *testing.T) {
	names, err := StoredUnitInventoryTypeNames(packageSourceDir())
	if err != nil {
		t.Fatalf("StoredUnitInventoryTypeNames: %v", err)
	}
	if len(names) != 1 {
		t.Fatalf("found %d stored-unit-inventory-shaped types %v, want exactly 1 (CQ-007/FR-105)", len(names), names)
	}
	if names[0] != "SegmentTableSlot" {
		t.Fatalf("the sole stored-unit inventory type is %q, want %q", names[0], "SegmentTableSlot")
	}
}

// TestStoredUnitInventoryAuditDetectsASecondCandidate proves the audit is
// not vacuously true (e.g. a bug that always returns 0 or 1 matches
// regardless of input): given synthetic source declaring two independent
// types with the inventory shape, it must report both.
func TestStoredUnitInventoryAuditDetectsASecondCandidate(t *testing.T) {
	const src = `package fakecontainer

import "Protodoc/pkg/pdlfmt"

type FirstInventory struct {
	UnitType byte
	Offset   uint64
	Digest   pdlfmt.Digest256
}

type SecondInventory struct {
	EntryType byte
	Offset    uint64
	Digest    pdlfmt.Digest256
}

type NotAnInventory struct {
	Name string
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "fake.go", src, 0)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	var names []string
	for _, decl := range f.Decls {
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
	if len(names) != 2 {
		t.Fatalf("synthetic fixture: got %d matches %v, want 2 (FirstInventory, SecondInventory)", len(names), names)
	}
}

// TestStoredUnitInventoryAuditIgnoresDerivedViewTypes confirms a read-only
// projection type such as StoredUnit (T-0021's TR-006 enumeration result,
// which names a unit's type/length/digest but not its file offset, and so
// cannot serve as an alternative locating index) is not itself flagged as
// a second inventory: it is a derived view over the one inventory, not an
// independently maintained second one.
func TestStoredUnitInventoryAuditIgnoresDerivedViewTypes(t *testing.T) {
	names, err := StoredUnitInventoryTypeNames(packageSourceDir())
	if err != nil {
		t.Fatalf("StoredUnitInventoryTypeNames: %v", err)
	}
	for _, n := range names {
		if n == "StoredUnit" {
			t.Fatalf("StoredUnit (a derived enumeration view, not a second inventory) was flagged as one")
		}
	}
}
