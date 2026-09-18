package semantics

import "testing"

// TestConformanceTBL001HeaderScope is T-0282's named conformance test (vector
// CONFORMANCE-TBL-001-header-scope; FR-039). Rule PD-TBL-001a requires every
// header cell to declare a scope that resolves to at least one real data cell;
// a header heading no data cell is rejected naming the header cell.
func TestConformanceTBL001HeaderScope(t *testing.T) {
	r0, r1 := pdUnitSem(0x10), pdUnitSem(0x11)
	c0, c1 := pdUnitSem(0x20), pdUnitSem(0x21)

	// A well-formed table: column header at (r0,c1) heads the data cell (r1,c1).
	good := Table{
		Rows:    unitList(0x10, 0x11),
		Columns: unitList(0x20, 0x21),
		Cells: []TableCell{
			{Row: r0, Col: c1, Kind: CellHeader, Scope: ScopeColumn},
			{Row: r1, Col: c1, Kind: CellData},
			{Row: r0, Col: c0, Kind: CellData},
			{Row: r1, Col: c0, Kind: CellData},
		},
	}
	if f := CheckHeaderScope(good); len(f) != 0 {
		t.Errorf("well-scoped header should pass, got %+v", f)
	}

	// A header whose column has no data cell: header at (r0,c1) scope=column,
	// but no data cell exists in column c1.
	orphan := Table{
		Rows:    unitList(0x10, 0x11),
		Columns: unitList(0x20, 0x21),
		Cells: []TableCell{
			{Row: r0, Col: c1, Kind: CellHeader, Scope: ScopeColumn},
			{Row: r0, Col: c0, Kind: CellData},
			{Row: r1, Col: c0, Kind: CellData},
		},
	}
	f := CheckHeaderScope(orphan)
	if len(f) != 1 || f[0].HeaderCol != c1 || f[0].Rule != RuleHeaderScope {
		t.Fatalf("orphan header: findings = %+v, want one PD-TBL-001a naming col c1", f)
	}
}
