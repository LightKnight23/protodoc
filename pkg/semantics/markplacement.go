// Renderable-mark placement (FR-037, PD-A11Y-001; T-0277). A renderable mark
// is any glyph a reader sees that is not itself a content unit's text — a
// footnote/endnote reference mark, a numbering label, a decorative rule or
// ornament. For accessibility, every such mark MUST declare where it sits in
// the logical reading order: either it references a ROOT_SEQUENCE position (it
// is read at that point in the reading order, e.g. a footnote reference mark)
// or it is flagged as pure decoration (it carries no reading-order meaning and
// assistive technology skips it). A mark that declares neither is ambiguous to
// assistive technology and is rejected by rule PD-A11Y-001.
//
// Provisional per T-0267 (clarify-002.md, OPEN awaiting Eyvar).
package semantics

import "Protodoc/pkg/pdlfmt"

// RuleMarkPlacement is the validator rule id requiring a renderable mark to
// declare a reading-order placement or be flagged decorative.
const RuleMarkPlacement = "PD-A11Y-001"

// RenderableMark is one mark record. Exactly one of the two placement signals
// must be present:
//   - PositionRef set (PositionRefPresent true): the mark is read at the
//     referenced unit's ROOT_SEQUENCE position.
//   - Decoration true: the mark is pure decoration, skipped by AT.
//
// A mark with neither (nor both) fails PD-A11Y-001.
type RenderableMark struct {
	UnitID pdlfmt.UnitID
	// PositionRef is the unit-id whose ROOT_SEQUENCE position fixes this
	// mark's reading-order placement; meaningful only when PositionRefPresent.
	PositionRef        pdlfmt.UnitID
	PositionRefPresent bool
	// Decoration flags the mark as carrying no reading-order meaning.
	Decoration bool
}

// MarkPlacementFinding is a PD-A11Y-001 rejection naming the offending mark.
type MarkPlacementFinding struct {
	Rule   string
	UnitID pdlfmt.UnitID
	Reason string
}

// CheckMarkPlacement applies PD-A11Y-001 to a set of renderable marks. A mark
// MUST declare exactly one placement signal: either a ROOT_SEQUENCE position
// reference or the decoration flag. A mark with neither is rejected (ambiguous
// to assistive technology); a mark asserting both is also rejected
// (contradictory). Findings name the offending mark's unit id. An empty result
// means every mark is well-placed.
func CheckMarkPlacement(marks []RenderableMark) []MarkPlacementFinding {
	var findings []MarkPlacementFinding
	for _, m := range marks {
		switch {
		case !m.PositionRefPresent && !m.Decoration:
			findings = append(findings, MarkPlacementFinding{
				Rule: RuleMarkPlacement, UnitID: m.UnitID,
				Reason: "mark declares neither a ROOT_SEQUENCE position reference nor the decoration flag",
			})
		case m.PositionRefPresent && m.Decoration:
			findings = append(findings, MarkPlacementFinding{
				Rule: RuleMarkPlacement, UnitID: m.UnitID,
				Reason: "mark declares both a position reference and the decoration flag (contradictory)",
			})
		}
	}
	return findings
}
