package semantics

import "testing"

// TestConformanceA11Y001MarkPlacement is T-0277's named conformance test
// (vector CONFORMANCE-A11Y-001-mark-placement; FR-037). Rule PD-A11Y-001
// requires every renderable mark to carry either a ROOT_SEQUENCE position
// reference or an explicit decoration flag; a mark with neither (or both) is
// rejected naming the mark's unit id.
func TestConformanceA11Y001MarkPlacement(t *testing.T) {
	positioned := RenderableMark{UnitID: pdUnitSem(0x01), PositionRef: pdUnitSem(0x40), PositionRefPresent: true}
	decorative := RenderableMark{UnitID: pdUnitSem(0x02), Decoration: true}
	neither := RenderableMark{UnitID: pdUnitSem(0x03)}
	both := RenderableMark{UnitID: pdUnitSem(0x04), PositionRef: pdUnitSem(0x41), PositionRefPresent: true, Decoration: true}

	// A well-placed pair passes.
	if f := CheckMarkPlacement([]RenderableMark{positioned, decorative}); len(f) != 0 {
		t.Errorf("well-placed marks should pass, got %+v", f)
	}

	// Neither signal -> rejected naming 0x03.
	f := CheckMarkPlacement([]RenderableMark{positioned, neither, decorative})
	if len(f) != 1 || f[0].UnitID != pdUnitSem(0x03) || f[0].Rule != RuleMarkPlacement {
		t.Errorf("neither-signal: findings = %+v, want one PD-A11Y-001 naming 0x03", f)
	}

	// Both signals -> rejected naming 0x04 (contradictory).
	fb := CheckMarkPlacement([]RenderableMark{both})
	if len(fb) != 1 || fb[0].UnitID != pdUnitSem(0x04) {
		t.Errorf("both-signal: findings = %+v, want one rejection naming 0x04", fb)
	}
}
