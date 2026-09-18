// Table header-scope validation (FR-039; T-0282). Validator rule PD-TBL-001a:
// every header cell must carry a scope, and that scope must resolve to at
// least one real data cell in the table's grid. A header heading no data cell
// is a structural reject naming the header cell.
//
// Provisional per T-0267 (clarify-002.md, OPEN awaiting Eyvar).
package semantics

import "Protodoc/pkg/pdlfmt"

// RuleHeaderScope is the validator rule id for header-cell scope resolution.
const RuleHeaderScope = "PD-TBL-001a"

// Table is the minimal grid model the scope/tiling rules need: the row and
// column id sequences (in display order) and the cell entries.
type Table struct {
	Rows    []pdlfmt.UnitID
	Columns []pdlfmt.UnitID
	Cells   []TableCell
}

// HeaderScopeFinding is a PD-TBL-001a rejection naming a header cell that
// heads no real data cell.
type HeaderScopeFinding struct {
	Rule      string
	HeaderRow pdlfmt.UnitID
	HeaderCol pdlfmt.UnitID
	Reason    string
}

// dataCellsHeadedBy returns how many DATA cells a header cell at (hRow,hCol)
// with the given scope heads.
//   - ScopeRow / ScopeRowGroup: data cells sharing the header's row.
//   - ScopeColumn / ScopeColumnGroup: data cells sharing the header's column.
//
// (Group scopes head at least their own row/column band; the "at least one"
// resolution check below is satisfied identically for band and single scope.)
func dataCellsHeadedBy(tbl Table, hRow, hCol pdlfmt.UnitID, scope CellScope) int {
	n := 0
	for _, c := range tbl.Cells {
		if c.Kind != CellData {
			continue
		}
		switch scope {
		case ScopeRow, ScopeRowGroup:
			if c.Row == hRow {
				n++
			}
		case ScopeColumn, ScopeColumnGroup:
			if c.Col == hCol {
				n++
			}
		}
	}
	return n
}

// CheckHeaderScope applies PD-TBL-001a: every header cell must declare a valid
// scope that resolves to at least one real data cell in the grid. A header
// with an invalid scope value, or one heading no data cell, is reported naming
// the header cell. An empty result means every header is well-scoped.
func CheckHeaderScope(tbl Table) []HeaderScopeFinding {
	var findings []HeaderScopeFinding
	for _, c := range tbl.Cells {
		if c.Kind != CellHeader {
			continue
		}
		if !c.Scope.ValidScope() {
			findings = append(findings, HeaderScopeFinding{
				Rule: RuleHeaderScope, HeaderRow: c.Row, HeaderCol: c.Col,
				Reason: "header cell declares an out-of-set scope value",
			})
			continue
		}
		if dataCellsHeadedBy(tbl, c.Row, c.Col, c.Scope) == 0 {
			findings = append(findings, HeaderScopeFinding{
				Rule: RuleHeaderScope, HeaderRow: c.Row, HeaderCol: c.Col,
				Reason: "header cell scope resolves to no real data cell in the grid",
			})
		}
	}
	return findings
}
