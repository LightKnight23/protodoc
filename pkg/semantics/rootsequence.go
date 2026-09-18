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
