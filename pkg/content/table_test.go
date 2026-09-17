package content

import (
	"bytes"
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// mkTable builds an M-row x N-column table fully tiled with one cell per
// (row, col) pair, returning the table and its row/col id slices.
func mkTable(t *testing.T, m, n int) Table {
	t.Helper()
	id, _ := testMintID()
	tbl := Table{ID: id}
	for i := 0; i < m; i++ {
		rid, _ := testMintID()
		tbl.Rows = append(tbl.Rows, rid)
	}
	for j := 0; j < n; j++ {
		cid, _ := testMintID()
		tbl.Columns = append(tbl.Columns, cid)
	}
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			content, _ := testMintID()
			tbl.Cells = append(tbl.Cells, CellEntry{Row: tbl.Rows[i], Col: tbl.Columns[j], Content: content})
		}
	}
	return tbl
}

// TestFR_082_TableGridTilesExactlyOnce is T-0367's named test. A fully tiled
// M x N table round-trips byte-exact and passes tiling validation; a cell
// naming an absent row/column, a duplicated (row, col) pair, and an
// uncovered (row, col) pair are each rejected naming the offending cell.
func TestFR_082_TableGridTilesExactlyOnce(t *testing.T) {
	tbl := mkTable(t, 3, 4)

	// Byte-exact round trip.
	enc := EncodeTable(nil, tbl)
	got, n, err := DecodeTable(enc)
	if err != nil {
		t.Fatalf("DecodeTable: %v", err)
	}
	if n != len(enc) {
		t.Fatalf("DecodeTable consumed %d of %d octets", n, len(enc))
	}
	reEnc := EncodeTable(nil, got)
	if !bytes.Equal(enc, reEnc) {
		t.Fatalf("table did not round-trip byte-exact")
	}

	// A correctly tiled grid validates.
	if err := tbl.ValidateTiling(); err != nil {
		t.Fatalf("fully tiled table failed validation: %v", err)
	}

	// Cell naming an absent row is rejected naming the cell.
	absentRow, _ := testMintID()
	bad := cloneTable(tbl)
	bad.Cells[0].Row = absentRow
	if err := bad.ValidateTiling(); !errors.Is(err, ErrTableCellUnknownRow) {
		t.Fatalf("absent-row cell: got %v, want ErrTableCellUnknownRow", err)
	}

	// Cell naming an absent column is rejected.
	absentCol, _ := testMintID()
	bad = cloneTable(tbl)
	bad.Cells[0].Col = absentCol
	if err := bad.ValidateTiling(); !errors.Is(err, ErrTableCellUnknownColumn) {
		t.Fatalf("absent-col cell: got %v, want ErrTableCellUnknownColumn", err)
	}

	// Duplicated (row, col) pair is rejected, and the error names the pair.
	bad = cloneTable(tbl)
	bad.Cells[1].Row = bad.Cells[0].Row
	bad.Cells[1].Col = bad.Cells[0].Col
	err = bad.ValidateTiling()
	if !errors.Is(err, ErrTableDuplicateCell) {
		t.Fatalf("duplicate cell: got %v, want ErrTableDuplicateCell", err)
	}
	var te *TableTilingError
	if !errors.As(err, &te) || !te.Row.Equal(bad.Cells[0].Row) || !te.Col.Equal(bad.Cells[0].Col) {
		t.Fatalf("duplicate-cell error did not name the offending (row,col): %v", err)
	}

	// Uncovered (row, col) pair: drop a cell, leaving a gap.
	bad = cloneTable(tbl)
	bad.Cells = bad.Cells[:len(bad.Cells)-1]
	if err := bad.ValidateTiling(); !errors.Is(err, ErrTableUncoveredCell) {
		t.Fatalf("uncovered pair: got %v, want ErrTableUncoveredCell", err)
	}
}

func cloneTable(t Table) Table {
	c := Table{ID: t.ID}
	c.Rows = append([]pdlfmt.UnitID(nil), t.Rows...)
	c.Columns = append([]pdlfmt.UnitID(nil), t.Columns...)
	c.Cells = append([]CellEntry(nil), t.Cells...)
	return c
}
