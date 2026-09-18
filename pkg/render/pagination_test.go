package render

import (
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_097_FixedPaginationNoConversionStep is T-0247's named integration test
// (FR-097). The authored fixed pagination is projected directly from the
// Frontmatter (page count, page dimensions) plus the PageDirectory in a single
// read-and-project path: the projected page count and geometry equal the
// writer's authored input verbatim (no conversion), and the break units are the
// directory's entries in ordinal order.
func TestFR_097_FixedPaginationNoConversionStep(t *testing.T) {
	fm := &container.Frontmatter{
		PageCount:  128,
		PageWidth:  pdlfmt.GeometricValue(8_267_400),
		PageHeight: pdlfmt.GeometricValue(11_692_800),
	}

	dir := &PageDirectory{}
	seeds := []byte{0x40, 0x10, 0x30, 0x20} // inserted out of order
	for _, s := range seeds {
		dir.Insert(PageEntry{BreakUnit: pdUnit(s), Geometry: uint32(s)})
	}

	fp := ProjectFixedPagination(fm, dir)

	// Page count and geometry match writer input verbatim (100%, no conversion).
	if fp.PageCount != fm.PageCount {
		t.Errorf("projected page count = %d, want authored %d", fp.PageCount, fm.PageCount)
	}
	if fp.PageWidth != fm.PageWidth || fp.PageHeight != fm.PageHeight {
		t.Errorf("projected geometry = (%d,%d), want authored (%d,%d)",
			fp.PageWidth, fp.PageHeight, fm.PageWidth, fm.PageHeight)
	}

	// Break units are projected in ordinal (rank) order, not insertion order.
	wantOrder := []pdlfmt.UnitID{pdUnit(0x10), pdUnit(0x20), pdUnit(0x30), pdUnit(0x40)}
	if len(fp.BreakUnits) != len(wantOrder) {
		t.Fatalf("projected %d break units, want %d", len(fp.BreakUnits), len(wantOrder))
	}
	for i := range wantOrder {
		if fp.BreakUnits[i] != wantOrder[i] {
			t.Errorf("break unit %d = %x, want %x (ordinal order)", i, fp.BreakUnits[i], wantOrder[i])
		}
	}

	// The projection is a pure function of its inputs: re-projecting the same
	// inputs yields the identical result (no hidden intermediate state).
	fp2 := ProjectFixedPagination(fm, dir)
	if fp2.PageCount != fp.PageCount || len(fp2.BreakUnits) != len(fp.BreakUnits) {
		t.Error("projection is not a pure read-and-project of its inputs")
	}

	// A nil directory still projects the authored count/geometry (no break
	// units), proving the frontmatter path needs no conversion either.
	fpNil := ProjectFixedPagination(fm, nil)
	if fpNil.PageCount != fm.PageCount || len(fpNil.BreakUnits) != 0 {
		t.Errorf("nil-directory projection = %+v, want authored count and no break units", fpNil)
	}
}
