// Table cell tiling validation (FR-082; T-0283). Validator rule PD-TBL-001b,
// the tiling half of the base rule PD-TBL-001 (spec.md FR-082 "Verify"): a
// table's cells MUST tile its declared tbl_rows x tbl_columns grid EXACTLY
// once — no two cells occupy the same (row, col) position (overlap), and no
// in-range (row, col) position is left uncovered (gap). A violation is
// reported naming the offending coordinate; zero false accepts, no repair.
//
// Provisional per T-0267 (clarify-002.md, OPEN awaiting Eyvar).
package semantics

import "Protodoc/pkg/pdlfmt"

// RuleCellTiling is the validator rule id for the tiling invariant. It is the
// concrete realization of the base rule PD-TBL-001 for the tiling half.
const RuleCellTiling = "PD-TBL-001b"

// TilingViolationKind classifies a tiling defect.
type TilingViolationKind uint8

const (
	// TilingOverlap: two cells name the same (row, col).
	TilingOverlap TilingViolationKind = iota
	// TilingGap: an in-range (row, col) is uncovered.
	TilingGap
)

// TilingFinding is a PD-TBL-001b rejection naming the offending grid position.
type TilingFinding struct {
	Rule string
	Row  pdlfmt.UnitID
	Col  pdlfmt.UnitID
	Kind TilingViolationKind
}

// CheckCellTiling applies the tiling invariant (PD-TBL-001b / base PD-TBL-001):
// the cells must cover every (row, col) in the tbl_rows x tbl_columns grid
// exactly once. It reports every overlap (first the duplicated coordinate) and
// every gap (uncovered coordinate), naming the coordinate. An empty result
// means the grid tiles exactly once.
func CheckCellTiling(tbl Table) []TilingFinding {
	var findings []TilingFinding

	type coord struct{ r, c pdlfmt.UnitID }
	count := make(map[coord]int, len(tbl.Cells))
	for _, cell := range tbl.Cells {
		count[coord{cell.Row, cell.Col}]++
	}

	// Overlaps: any coordinate covered more than once, reported once.
	reported := make(map[coord]bool)
	for _, cell := range tbl.Cells {
		k := coord{cell.Row, cell.Col}
		if count[k] > 1 && !reported[k] {
			findings = append(findings, TilingFinding{Rule: RuleCellTiling, Row: k.r, Col: k.c, Kind: TilingOverlap})
			reported[k] = true
		}
	}

	// Gaps: every (row, col) in range must be covered at least once.
	for _, r := range tbl.Rows {
		for _, c := range tbl.Columns {
			if count[coord{r, c}] == 0 {
				findings = append(findings, TilingFinding{Rule: RuleCellTiling, Row: r, Col: c, Kind: TilingGap})
			}
		}
	}
	return findings
}
