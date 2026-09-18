// PD-BIDI-001 missing-direction validator rule (T-0270; FR-032). A text-
// container record whose mandatory direction field is absent is rejected,
// naming the offending unit id and the octet offset (the standard finding
// shape used across the validator). Provisional per the T-0267 ruling
// (clarify-002.md, OPEN).
package semantics

import (
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// RuleMissingDirection is the validator rule id for a text-container record
// missing its mandatory direction field.
const RuleMissingDirection = "PD-BIDI-001"

// DirectionFinding is a PD-BIDI-001 rejection finding, carrying the rule id, the
// offending unit id, and the octet offset of the record (standard finding
// shape).
type DirectionFinding struct {
	Rule     string
	UnitID   pdlfmt.UnitID
	OctetOff uint64
}

func (f DirectionFinding) String() string {
	return fmt.Sprintf("%s: text-container unit %x at octet %d has no mandatory direction field",
		f.Rule, f.UnitID, f.OctetOff)
}

// TextContainerRecord is the decode-time view of a text-container relevant to
// this rule: its unit id, its octet offset in the stream, and whether the
// mandatory direction field was present with a valid value.
type TextContainerRecord struct {
	UnitID           pdlfmt.UnitID
	OctetOff         uint64
	DirectionPresent bool
	DirectionValue   Direction
}

// CheckDirectionPresent applies PD-BIDI-001: it returns a *DirectionFinding
// naming the unit id and octet offset when a text-container record's direction
// field is absent (or present but out of the closed value set), or nil when a
// valid direction is present.
func CheckDirectionPresent(rec TextContainerRecord) *DirectionFinding {
	if !rec.DirectionPresent || !rec.DirectionValue.Valid() {
		return &DirectionFinding{Rule: RuleMissingDirection, UnitID: rec.UnitID, OctetOff: rec.OctetOff}
	}
	return nil
}
