// Structural locator emission (T-0088, FR-042): the extraction view emits,
// per text unit, a locator of the unit's content identity plus a live scalar
// position. ScalarOffset is computed at extraction time by counting Unicode
// scalar values within the current unit only (CON-001); it is never read
// from a persisted counted-position field, because none exists. Importing
// the content package is within the extraction module boundary (content is
// not integrity/render/merge).
package extract

import (
	"Protodoc/pkg/content"
	"Protodoc/pkg/pdlfmt"
)

// Locator addresses an emitted text position by content identity plus a
// live, unit-scoped scalar offset (FR-042). UnitID is the run's identity;
// ScalarOffset is the count of Unicode scalar values from the start of the
// unit to this position, computed live, never persisted.
type Locator struct {
	UnitID       pdlfmt.UnitID
	ScalarOffset uint32
}

// LocatorsForUnit emits a locator for the start of each run in a text unit's
// run sequence, in reading order. ScalarOffset accumulates the scalar-value
// length of the preceding runs within THIS unit only, computed live from run
// text via the content package's scalar counting -- it resets per unit and
// is never sourced from a stored field. It returns one locator per run.
func LocatorsForUnit(runs []content.Run) []Locator {
	out := make([]Locator, 0, len(runs))
	var offset uint32
	for _, r := range runs {
		out = append(out, Locator{UnitID: r.RunID, ScalarOffset: offset})
		offset += uint32(r.ScalarLen())
	}
	return out
}

// LocatorAt returns the locator addressing the scalar-value at run-internal
// index i within run r: the run's identity and i as the (unit-scoped) scalar
// offset. It is the point-addressing form used when extracting a specific
// position rather than run starts.
func LocatorAt(r content.Run, i int) (Locator, bool) {
	if i < 0 || i >= r.ScalarLen() {
		return Locator{}, false
	}
	return Locator{UnitID: r.RunID, ScalarOffset: uint32(i)}, true
}
