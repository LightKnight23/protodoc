package semantics

import "testing"

// TestConformanceTBL002CellTiling is T-0283's named conformance test (vector
// CONFORMANCE-TBL-002-cell-tiling; FR-082). Rule PD-TBL-001b (the tiling half
// of the base rule PD-TBL-001) asserts a table's cells tile its declared grid
// exactly once: no overlap, no gap, naming the offending coordinate.
func TestConformanceTBL002CellTiling(t *testing.T) {
	r0, r1 := pdUnitSem(0x10), pdUnitSem(0x11)
	c0, c1 := pdUnitSem(0x20), pdUnitSem(0x21)
	rows := unitList(0x10, 0x11)
	cols := unitList(0x20, 0x21)

	// A complete 2x2 tiling passes.
	full := Table{Rows: rows, Columns: cols, Cells: []TableCell{
		{Row: r0, Col: c0}, {Row: r0, Col: c1},
		{Row: r1, Col: c0}, {Row: r1, Col: c1},
	}}
	if f := CheckCellTiling(full); len(f) != 0 {
		t.Errorf("complete tiling should pass, got %+v", f)
	}

	// A gap: (r1,c1) uncovered -> reported as a gap naming that coordinate.
	gap := Table{Rows: rows, Columns: cols, Cells: []TableCell{
		{Row: r0, Col: c0}, {Row: r0, Col: c1}, {Row: r1, Col: c0},
	}}
	fg := CheckCellTiling(gap)
	if len(fg) != 1 || fg[0].Kind != TilingGap || fg[0].Row != r1 || fg[0].Col != c1 {
		t.Fatalf("gap: findings = %+v, want one gap naming (r1,c1)", fg)
	}
	if fg[0].Rule != RuleCellTiling {
		t.Errorf("rule = %q, want %q", fg[0].Rule, RuleCellTiling)
	}

	// An overlap: (r0,c0) covered twice -> reported as overlap naming it.
	overlap := Table{Rows: rows, Columns: cols, Cells: []TableCell{
		{Row: r0, Col: c0}, {Row: r0, Col: c0}, {Row: r0, Col: c1},
		{Row: r1, Col: c0}, {Row: r1, Col: c1},
	}}
	fo := CheckCellTiling(overlap)
	foundOverlap := false
	for _, x := range fo {
		if x.Kind == TilingOverlap && x.Row == r0 && x.Col == c0 {
			foundOverlap = true
		}
	}
	if !foundOverlap {
		t.Errorf("overlap: findings = %+v, want an overlap naming (r0,c0)", fo)
	}
}
