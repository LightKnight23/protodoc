// Numbering-literal rejection (FR-083; T-0287). Validator rule PD-A11Y-004: an
// ordered-sequence item MUST NOT carry a persisted literal numbering label
// that is not derivable from its SEQUENCE_DEFINITION and order-value. A literal
// that disagrees with the derived label is a structural reject — the derived
// label is the only authoritative one. (A persisted literal that exactly
// matches the derived label is redundant but harmless; the rule targets
// authoritative literals that would override derivation.)
//
// Provisional per T-0267 (clarify-002.md, OPEN awaiting Eyvar).
package semantics

import "Protodoc/pkg/pdlfmt"

// RuleNumberingLiteral is the validator rule id for numbering-literal rejection.
const RuleNumberingLiteral = "PD-A11Y-004"

// LabeledOrderedItem is an ordered item that (illegally) also persists a
// literal label. PersistedLabel is present only when HasPersistedLabel is true.
type LabeledOrderedItem struct {
	Item              OrderedItem
	HasPersistedLabel bool
	PersistedLabel    string
}

// NumberingLiteralFinding is a PD-A11Y-004 rejection naming the offending item.
type NumberingLiteralFinding struct {
	Rule           string
	ItemID         pdlfmt.UnitID
	PersistedLabel string
	DerivedLabel   string
}

// CheckNumberingLiteral applies PD-A11Y-004. For each labeled item it derives
// the authoritative label from the item's SEQUENCE_DEFINITION (looked up in
// defs by def_id) and order-value, and rejects the item if it persists a
// literal label that differs from the derived one (an authoritative literal
// that would override derivation). rootSeqPos maps an item id to its
// ROOT_SEQUENCE position. An item whose def_id is unknown is rejected too (its
// literal cannot be shown derivable). An empty result means no illegal literal.
func CheckNumberingLiteral(items []LabeledOrderedItem, defs map[pdlfmt.UnitID]SequenceDefinition, rootSeqPos map[pdlfmt.UnitID]int) []NumberingLiteralFinding {
	var findings []NumberingLiteralFinding
	for _, li := range items {
		if !li.HasPersistedLabel {
			continue // no literal persisted: nothing to reject
		}
		def, ok := defs[li.Item.DefID]
		if !ok {
			findings = append(findings, NumberingLiteralFinding{
				Rule: RuleNumberingLiteral, ItemID: li.Item.ItemID,
				PersistedLabel: li.PersistedLabel, DerivedLabel: "",
			})
			continue
		}
		derived := DeriveNumberingLabel(def, li.Item, rootSeqPos[li.Item.ItemID]).Value
		if li.PersistedLabel != derived {
			findings = append(findings, NumberingLiteralFinding{
				Rule: RuleNumberingLiteral, ItemID: li.Item.ItemID,
				PersistedLabel: li.PersistedLabel, DerivedLabel: derived,
			})
		}
	}
	return findings
}
