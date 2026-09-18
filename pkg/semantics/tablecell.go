// Table cell kind + scope (FR-039; T-0281..T-0283). A cell is either a data
// cell or a header cell; a header cell declares the accessibility scope it
// heads (row / column / row-group / column-group). This mirrors document.abnf
// S4's cell-kind / cell-scope fields and data-model.md 2.24.
//
// Provisional per T-0267 (clarify-002.md, OPEN awaiting Eyvar).
package semantics

import (
	"errors"

	"Protodoc/pkg/pdlfmt"
)

// CellKind is the data/header discriminant of a table cell.
type CellKind uint8

const (
	// CellData is a plain data cell.
	CellData CellKind = 0x00
	// CellHeader is a header cell (heads some data cells).
	CellHeader CellKind = 0x01
)

// CellScope is the accessibility scope a header cell heads.
type CellScope uint8

const (
	// ScopeRow: the header heads its row's data cells.
	ScopeRow CellScope = 0x00
	// ScopeColumn: the header heads its column's data cells.
	ScopeColumn CellScope = 0x01
	// ScopeRowGroup: the header heads a group of rows.
	ScopeRowGroup CellScope = 0x02
	// ScopeColumnGroup: the header heads a group of columns.
	ScopeColumnGroup CellScope = 0x03
)

// ErrCellKindOutOfSet / ErrCellScopeOutOfSet reject reserved values.
var (
	ErrCellKindOutOfSet  = errors.New("semantics: cell-kind outside closed value set {0,1}")
	ErrCellScopeOutOfSet = errors.New("semantics: cell-scope outside closed value set {0,1,2,3}")
)

// ValidKind reports whether k is an assigned cell-kind value.
func (k CellKind) ValidKind() bool { return k == CellData || k == CellHeader }

// ValidScope reports whether s is an assigned cell-scope value.
func (s CellScope) ValidScope() bool { return s <= ScopeColumnGroup }

// TableCell is one cell-entry.
type TableCell struct {
	Row     pdlfmt.UnitID
	Col     pdlfmt.UnitID
	Content pdlfmt.UnitID
	Kind    CellKind
	Scope   CellScope
}

// EncodeCellKindScope appends the cell-kind and cell-scope octets (the two
// fields T-0281 adds) to dst, after validating both against their closed sets.
func EncodeCellKindScope(dst []byte, kind CellKind, scope CellScope) ([]byte, error) {
	if !kind.ValidKind() {
		return nil, ErrCellKindOutOfSet
	}
	if !scope.ValidScope() {
		return nil, ErrCellScopeOutOfSet
	}
	return append(dst, byte(kind), byte(scope)), nil
}

// DecodeCellKindScope reads the cell-kind and cell-scope octets from buf,
// rejecting a reserved value in either. It returns the two values and the
// number of octets consumed (2).
func DecodeCellKindScope(buf []byte) (CellKind, CellScope, int, error) {
	if len(buf) < 2 {
		return 0, 0, 0, errors.New("semantics: short cell-kind/cell-scope buffer")
	}
	kind := CellKind(buf[0])
	scope := CellScope(buf[1])
	if !kind.ValidKind() {
		return 0, 0, 0, ErrCellKindOutOfSet
	}
	if !scope.ValidScope() {
		return 0, 0, 0, ErrCellScopeOutOfSet
	}
	return kind, scope, 2, nil
}
