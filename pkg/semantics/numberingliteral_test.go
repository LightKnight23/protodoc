package semantics

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestConformanceA11Y004NumberingLiteralRejected is T-0287's named conformance
// test (vector CONFORMANCE-A11Y-004-numbering-literal-rejected; FR-083). Rule
// PD-A11Y-004 rejects an ordered item carrying a persisted literal numbering
// label not derivable from its SEQUENCE_DEFINITION and order-value.
func TestConformanceA11Y004NumberingLiteralRejected(t *testing.T) {
	def := SequenceDefinition{DefID: pdUnitSem(0x01), Style: StyleDecimal, Start: 1, Suffix: "."}
	defs := map[pdlfmt.UnitID]SequenceDefinition{def.DefID: def}
	pos := map[pdlfmt.UnitID]int{}

	item := OrderedItem{ItemID: pdUnitSem(0x10), DefID: def.DefID, OrderValue: 2} // derives "3."

	// No persisted literal: passes.
	if f := CheckNumberingLiteral([]LabeledOrderedItem{{Item: item}}, defs, pos); len(f) != 0 {
		t.Errorf("item without a persisted literal should pass, got %+v", f)
	}

	// Persisted literal disagreeing with the derived label: rejected.
	bad := LabeledOrderedItem{Item: item, HasPersistedLabel: true, PersistedLabel: "99."}
	f := CheckNumberingLiteral([]LabeledOrderedItem{bad}, defs, pos)
	if len(f) != 1 || f[0].ItemID != item.ItemID || f[0].Rule != RuleNumberingLiteral {
		t.Fatalf("disagreeing literal: findings = %+v, want one PD-A11Y-004 naming the item", f)
	}
	if f[0].DerivedLabel != "3." || f[0].PersistedLabel != "99." {
		t.Errorf("finding should report derived %q vs persisted %q, got %+v", "3.", "99.", f[0])
	}

	// Unknown definition: the literal cannot be shown derivable -> rejected.
	orphan := LabeledOrderedItem{Item: OrderedItem{ItemID: pdUnitSem(0x11), DefID: pdUnitSem(0xEE), OrderValue: 0}, HasPersistedLabel: true, PersistedLabel: "1."}
	fo := CheckNumberingLiteral([]LabeledOrderedItem{orphan}, defs, pos)
	if len(fo) != 1 || fo[0].ItemID != pdUnitSem(0x11) {
		t.Errorf("unknown-def literal: findings = %+v, want one naming the item", fo)
	}
}
