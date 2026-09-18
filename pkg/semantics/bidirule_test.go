package semantics

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestConformanceBIDI001MissingDirectionRejected is T-0270's named conformance
// test (vector id CONFORMANCE-BIDI-001-missing-direction-rejected; FR-032). A
// text-container record with the direction field omitted is rejected by rule
// PD-BIDI-001, naming the correct unit id and octet offset; a record with a
// valid direction passes.
func TestConformanceBIDI001MissingDirectionRejected(t *testing.T) {
	unit := pdUnitSem(0x42)

	// Direction omitted -> rejected with the rule id, unit id, and offset.
	missing := TextContainerRecord{UnitID: unit, OctetOff: 4096, DirectionPresent: false}
	f := CheckDirectionPresent(missing)
	if f == nil {
		t.Fatal("a record with no direction field must be rejected")
	}
	if f.Rule != RuleMissingDirection {
		t.Errorf("rule = %q, want %q", f.Rule, RuleMissingDirection)
	}
	if f.UnitID != unit {
		t.Errorf("finding names unit %x, want %x", f.UnitID, unit)
	}
	if f.OctetOff != 4096 {
		t.Errorf("finding octet offset = %d, want 4096", f.OctetOff)
	}

	// Present but out-of-set -> also rejected.
	badVal := TextContainerRecord{UnitID: unit, OctetOff: 10, DirectionPresent: true, DirectionValue: Direction(9)}
	if CheckDirectionPresent(badVal) == nil {
		t.Error("an out-of-set direction value must be rejected")
	}

	// Valid direction present -> passes.
	ok := TextContainerRecord{UnitID: unit, OctetOff: 0, DirectionPresent: true, DirectionValue: DirectionRTL}
	if CheckDirectionPresent(ok) != nil {
		t.Error("a record with a valid direction must pass")
	}
}

// pdUnitSem builds a deterministic unit id from a seed for the semantics tests.
func pdUnitSem(seed byte) pdlfmt.UnitID {
	var id pdlfmt.UnitID
	id[0] = seed
	return id
}
