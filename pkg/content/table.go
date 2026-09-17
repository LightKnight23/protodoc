// TABLE base record (T-0367, FR-082; document.abnf S4, data-model.md S2.24).
// A table is a grid of cells that tiles exactly once: every cell resolves
// into the table's rows and columns, no two cells name the same (row, col)
// pair, and every (row, col) pair in range is named by exactly one cell. Row
// and column ids are content identities (minted once, reorderable like a
// run_id), never positional indices.
package content

import (
	"encoding/binary"
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// CellEntry binds a cell to its (row, col) identities and its content block
// (document.abnf S4 cell-entry). Row/Col are unit-ids that must resolve into
// the table's Rows/Columns; Content is the unit-id of the cell's TextBlock.
type CellEntry struct {
	Row     pdlfmt.UnitID
	Col     pdlfmt.UnitID
	Content pdlfmt.UnitID
}

// Table is the decoded TABLE record (document.abnf S4).
type Table struct {
	ID      pdlfmt.UnitID
	Rows    []pdlfmt.UnitID
	Columns []pdlfmt.UnitID
	Cells   []CellEntry
}

var (
	// ErrTableCellUnknownRow is returned when a cell names a row absent from
	// the table's Rows (FR-108: rejected, never clamped).
	ErrTableCellUnknownRow = errors.New("content: table cell names a row not present in tbl-rows")
	// ErrTableCellUnknownColumn is symmetric for columns.
	ErrTableCellUnknownColumn = errors.New("content: table cell names a column not present in tbl-columns")
	// ErrTableDuplicateCell is returned when two cells name the same
	// (row, col) pair (the grid would not tile exactly once).
	ErrTableDuplicateCell = errors.New("content: two table cells name the same (row, col) pair")
	// ErrTableUncoveredCell is returned when some (row, col) pair in range
	// is named by no cell (a gap in the tiling).
	ErrTableUncoveredCell = errors.New("content: a (row, col) pair in range is covered by no table cell")
	// ErrTableTruncated is returned when the encoding ends prematurely.
	ErrTableTruncated = errors.New("content: table encoding truncated")
)

// TableTilingError names the offending (row, col) pair for a tiling failure.
type TableTilingError struct {
	Kind error // one of the Err* tiling errors above
	Row  pdlfmt.UnitID
	Col  pdlfmt.UnitID
}

func (e *TableTilingError) Error() string {
	return fmt.Sprintf("%v: row %x col %x", e.Kind, e.Row, e.Col)
}

func (e *TableTilingError) Unwrap() error { return e.Kind }

// ValidateTiling checks the FR-082 grid-tiling invariant: every cell's row
// and column resolve into Rows/Columns, no (row, col) pair is named twice,
// and every (row, col) pair in the Rows x Columns range is named exactly
// once. It returns a *TableTilingError naming the offending pair on the
// first violation.
func (t Table) ValidateTiling() error {
	rowSet := idSet(t.Rows)
	colSet := idSet(t.Columns)

	type rc struct{ r, c pdlfmt.UnitID }
	named := make(map[rc]struct{}, len(t.Cells))
	for _, cell := range t.Cells {
		if _, ok := rowSet[cell.Row]; !ok {
			return &TableTilingError{Kind: ErrTableCellUnknownRow, Row: cell.Row, Col: cell.Col}
		}
		if _, ok := colSet[cell.Col]; !ok {
			return &TableTilingError{Kind: ErrTableCellUnknownColumn, Row: cell.Row, Col: cell.Col}
		}
		key := rc{cell.Row, cell.Col}
		if _, dup := named[key]; dup {
			return &TableTilingError{Kind: ErrTableDuplicateCell, Row: cell.Row, Col: cell.Col}
		}
		named[key] = struct{}{}
	}
	// Every (row, col) pair in range must be covered exactly once.
	for _, r := range t.Rows {
		for _, c := range t.Columns {
			if _, ok := named[rc{r, c}]; !ok {
				return &TableTilingError{Kind: ErrTableUncoveredCell, Row: r, Col: c}
			}
		}
	}
	return nil
}

func idSet(ids []pdlfmt.UnitID) map[pdlfmt.UnitID]struct{} {
	s := make(map[pdlfmt.UnitID]struct{}, len(ids))
	for _, id := range ids {
		s[id] = struct{}{}
	}
	return s
}

// EncodeTable appends a canonical encoding of t: id(16), then each of
// rows/columns/cells as a u32 count followed by the elements (a unit-id is
// 16 octets; a cell-entry is 48). The counts make the plain-seqs
// self-delimiting for byte-exact round trip.
func EncodeTable(dst []byte, t Table) []byte {
	dst = pdlfmt.AppendUnitID(dst, t.ID)
	dst = appendIDSeq(dst, t.Rows)
	dst = appendIDSeq(dst, t.Columns)
	var cnt [4]byte
	binary.BigEndian.PutUint32(cnt[:], uint32(len(t.Cells)))
	dst = append(dst, cnt[:]...)
	for _, c := range t.Cells {
		dst = pdlfmt.AppendUnitID(dst, c.Row)
		dst = pdlfmt.AppendUnitID(dst, c.Col)
		dst = pdlfmt.AppendUnitID(dst, c.Content)
	}
	return dst
}

func appendIDSeq(dst []byte, ids []pdlfmt.UnitID) []byte {
	var cnt [4]byte
	binary.BigEndian.PutUint32(cnt[:], uint32(len(ids)))
	dst = append(dst, cnt[:]...)
	for _, id := range ids {
		dst = pdlfmt.AppendUnitID(dst, id)
	}
	return dst
}

// DecodeTable decodes a Table from the leading octets of src, returning the
// table and the number of octets consumed. It verifies before allocating
// from a declared count (CP-006): a count implying more octets than remain
// is rejected as truncated rather than driving a huge allocation.
func DecodeTable(src []byte) (Table, int, error) {
	var t Table
	pos := 0
	if len(src) < 16 {
		return t, 0, ErrTableTruncated
	}
	id, _, err := pdlfmt.DecodeUnitID(src[:16])
	if err != nil {
		return t, 0, fmt.Errorf("content: tbl-id: %w", err)
	}
	t.ID = id
	pos += 16

	rows, n, err := decodeIDSeq(src[pos:])
	if err != nil {
		return t, 0, fmt.Errorf("content: tbl-rows: %w", err)
	}
	t.Rows = rows
	pos += n

	cols, n, err := decodeIDSeq(src[pos:])
	if err != nil {
		return t, 0, fmt.Errorf("content: tbl-columns: %w", err)
	}
	t.Columns = cols
	pos += n

	if len(src)-pos < 4 {
		return t, 0, ErrTableTruncated
	}
	cellCount := int(binary.BigEndian.Uint32(src[pos : pos+4]))
	pos += 4
	if cellCount < 0 || (len(src)-pos)/48 < cellCount {
		return t, 0, ErrTableTruncated
	}
	t.Cells = make([]CellEntry, cellCount)
	for i := 0; i < cellCount; i++ {
		r, _, err := pdlfmt.DecodeUnitID(src[pos : pos+16])
		if err != nil {
			return t, 0, fmt.Errorf("content: cell %d row: %w", i, err)
		}
		c, _, err := pdlfmt.DecodeUnitID(src[pos+16 : pos+32])
		if err != nil {
			return t, 0, fmt.Errorf("content: cell %d col: %w", i, err)
		}
		cc, _, err := pdlfmt.DecodeUnitID(src[pos+32 : pos+48])
		if err != nil {
			return t, 0, fmt.Errorf("content: cell %d content: %w", i, err)
		}
		t.Cells[i] = CellEntry{Row: r, Col: c, Content: cc}
		pos += 48
	}
	return t, pos, nil
}

func decodeIDSeq(src []byte) ([]pdlfmt.UnitID, int, error) {
	if len(src) < 4 {
		return nil, 0, ErrTableTruncated
	}
	count := int(binary.BigEndian.Uint32(src[:4]))
	if count < 0 || (len(src)-4)/16 < count {
		return nil, 0, ErrTableTruncated
	}
	ids := make([]pdlfmt.UnitID, count)
	pos := 4
	for i := 0; i < count; i++ {
		id, _, err := pdlfmt.DecodeUnitID(src[pos : pos+16])
		if err != nil {
			return nil, 0, err
		}
		ids[i] = id
		pos += 16
	}
	return ids, pos, nil
}
