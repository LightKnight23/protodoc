// ROOT_SEQUENCE authored reading-order record (T-0273; FR-036). The document's
// authoritative logical top-to-bottom reading order over its content units,
// independent of storage/append order. Provisional per the T-0267 ruling
// (clarify-002.md, OPEN).
package semantics

import "Protodoc/pkg/pdlfmt"

// RootSequence is the authored reading-order record: the singleton record's own
// identity plus the ordered list of every content unit, one entry each, in
// reading order (data-model.md 2.27).
type RootSequence struct {
	RSID  pdlfmt.UnitID
	Order []pdlfmt.UnitID
}

// Len returns the number of units in the reading order.
func (rs RootSequence) Len() int { return len(rs.Order) }

// Position returns the reading-order index of a unit and whether it is present.
func (rs RootSequence) Position(unit pdlfmt.UnitID) (int, bool) {
	for i, u := range rs.Order {
		if u == unit {
			return i, true
		}
	}
	return 0, false
}

// RuleRootSequenceCompleteness is the validator rule id requiring every content
// unit to appear in ROOT_SEQUENCE exactly once.
const RuleRootSequenceCompleteness = "PD-A11Y-005"

// RootSequenceViolationKind classifies a completeness violation.
type RootSequenceViolationKind uint8

const (
	// ViolationDuplicate: a unit appears in rs_order more than once.
	ViolationDuplicate RootSequenceViolationKind = iota
	// ViolationOmission: a content unit is absent from rs_order.
	ViolationOmission
)

// RootSequenceFinding is a PD-A11Y-005 rejection naming the rule, the offending
// unit id, and whether it was a duplicate or an omission.
type RootSequenceFinding struct {
	Rule   string
	UnitID pdlfmt.UnitID
	Kind   RootSequenceViolationKind
}

// CheckRootSequenceCompleteness applies PD-A11Y-005: it verifies rs_order lists
// every unit in `contentUnits` EXACTLY ONCE. It returns findings for every
// duplicate (a unit listed twice) and every omission (a content unit absent
// from rs_order), each naming the offending unit id. An empty result means the
// reading order is complete.
func (rs RootSequence) CheckRootSequenceCompleteness(contentUnits []pdlfmt.UnitID) []RootSequenceFinding {
	var findings []RootSequenceFinding

	seen := make(map[pdlfmt.UnitID]int, len(rs.Order))
	for _, u := range rs.Order {
		seen[u]++
	}
	// Duplicates (report each duplicated unit once, in rs_order first-seen order).
	reported := make(map[pdlfmt.UnitID]bool)
	for _, u := range rs.Order {
		if seen[u] > 1 && !reported[u] {
			findings = append(findings, RootSequenceFinding{Rule: RuleRootSequenceCompleteness, UnitID: u, Kind: ViolationDuplicate})
			reported[u] = true
		}
	}
	// Omissions: a content unit not present in rs_order.
	for _, u := range contentUnits {
		if seen[u] == 0 {
			findings = append(findings, RootSequenceFinding{Rule: RuleRootSequenceCompleteness, UnitID: u, Kind: ViolationOmission})
		}
	}
	return findings
}
