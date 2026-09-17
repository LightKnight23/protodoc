// Structural-failure diagnostics (T-0114, FR-102): for truncation,
// oversized-declared-length, and unparseable-TLV failures, a Finding
// carries the octet offset of the first divergence, the enclosing unit id
// (or an explicit NONE for a pre-unit prefix failure), and the violated
// rule id, in the shape the CLI report schema (a later M18/M19 task) expects.
package validate

import (
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// UnitRef identifies the enclosing content unit of a structural failure, or
// marks the failure as occurring in the pre-unit fixed prefix (None).
type UnitRef struct {
	None bool          // true: failure is in the pre-unit prefix (no enclosing unit)
	ID   pdlfmt.UnitID // the enclosing unit's identity when None is false
}

// NoUnit is the explicit "no enclosing unit" reference for a pre-unit
// prefix failure (FR-102 requires an explicit NONE rather than a zero id).
func NoUnit() UnitRef { return UnitRef{None: true} }

// InUnit references the enclosing unit id of a failure inside a content unit.
func InUnit(id pdlfmt.UnitID) UnitRef { return UnitRef{ID: id} }

// String renders the unit reference for a diagnostic ("NONE" or the hex id).
func (u UnitRef) String() string {
	if u.None {
		return "NONE"
	}
	return fmt.Sprintf("%x", u.ID)
}

// Rule ids for the three structural-failure classes this task diagnoses
// (contracts wire-grammar validator rules; FR-102).
const (
	// RuleTruncated: the declared structure runs past the available octets.
	RuleTruncated = "PD-TRUNC-001"
	// RuleOversizedLength: a declared length field exceeds its permitted bound.
	RuleOversizedLength = "PD-LEN-001"
	// RuleUnparseableTLV: a TLV field cannot be parsed (bad tag/length form).
	RuleUnparseableTLV = "PD-TLV-001"
)

// Diagnostic is the offset/unit/rule shape the CLI report schema (a later task) expects
// for a structural failure, carried inside a Finding.
type Diagnostic struct {
	Offset uint64  // octet offset of the first divergence
	Unit   UnitRef // enclosing unit, or NONE
	RuleID string  // the violated rule id
}

// StructuralFinding builds a validity Finding for a structural failure at the
// given step, with the offset/unit/rule diagnostic populated. The Finding's
// RuleID mirrors the diagnostic's rule id and its message names all three
// coordinates, so a caller reading the verdict alone learns where and why.
func StructuralFinding(step StepID, d Diagnostic) *Finding {
	return &Finding{
		Step:    step,
		RuleID:  d.RuleID,
		Message: fmt.Sprintf("structural failure at offset %d in unit %s: %s", d.Offset, d.Unit.String(), d.RuleID),
		Diag:    &d,
	}
}
