package semantics

import "testing"

// TestConformanceA11Y002RoleResolution is T-0280's named conformance test
// (vector CONFORMANCE-A11Y-002-role-resolution; FR-038). Rule PD-A11Y-002
// asserts every construct instance resolves to exactly one role; an unmapped
// or newly-added construct kind fails closed, rejected naming the construct.
func TestConformanceA11Y002RoleResolution(t *testing.T) {
	// All-mapped instances pass.
	ok := []ConstructInstance{
		{UnitID: pdUnitSem(0x01), Kind: KindTextBlock},
		{UnitID: pdUnitSem(0x02), Kind: KindTable},
		{UnitID: pdUnitSem(0x03), Kind: KindRootSequence},
	}
	if f := CheckRoleResolution(ok); len(f) != 0 {
		t.Errorf("all-mapped instances should resolve, got %+v", f)
	}

	// A reserved/unknown kind (0x2A, in the reserved 0x0F-0x3F range) fails
	// closed naming the construct.
	unknown := ConstructInstance{UnitID: pdUnitSem(0x09), Kind: ConstructKind(0x2A)}
	f := CheckRoleResolution([]ConstructInstance{ok[0], unknown})
	if len(f) != 1 || f[0].UnitID != pdUnitSem(0x09) || f[0].Kind != ConstructKind(0x2A) {
		t.Fatalf("unknown-kind: findings = %+v, want one naming 0x09 kind 0x2A", f)
	}
	if f[0].Rule != RuleRoleResolution {
		t.Errorf("rule = %q, want %q", f[0].Rule, RuleRoleResolution)
	}

	// Fail-closed: it is rejected, not silently assigned a role.
	if _, resolved := RoleOf(ConstructKind(0x2A)); resolved {
		t.Errorf("unmapped kind 0x2A must not resolve to any role")
	}
}
