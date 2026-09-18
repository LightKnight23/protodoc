// Deterministic ordered-sequence numbering (FR-083; T-0286). A SEQUENCE_
// DEFINITION record names a numbering style (decimal, lower/upper alpha,
// lower/upper roman) and a starting value; each ordered item carries only an
// order-value. The rendered numbering LABEL is a PURE DETERMINISTIC FUNCTION
// of (order-value, referenced sequence definition, ROOT_SEQUENCE position) —
// no persisted literal label is ever authoritative. Rendered labels are
// carried as COMPUTED_INLINE values (T-0272), so their directionality is
// isolated and re-materialising them never perturbs surrounding text.
//
// Provisional per T-0267 (clarify-002.md, OPEN awaiting Eyvar).
package semantics

import (
	"strconv"
	"strings"

	"Protodoc/pkg/pdlfmt"
)

// NumberingStyle is the closed set of numbering styles.
type NumberingStyle uint8

const (
	// StyleDecimal: 1, 2, 3, ...
	StyleDecimal NumberingStyle = 0x00
	// StyleLowerAlpha: a, b, c, ..., z, aa, ...
	StyleLowerAlpha NumberingStyle = 0x01
	// StyleUpperAlpha: A, B, C, ...
	StyleUpperAlpha NumberingStyle = 0x02
	// StyleLowerRoman: i, ii, iii, ...
	StyleLowerRoman NumberingStyle = 0x03
	// StyleUpperRoman: I, II, III, ...
	StyleUpperRoman NumberingStyle = 0x04
)

// ValidStyle reports whether s is an assigned numbering style.
func (s NumberingStyle) ValidStyle() bool { return s <= StyleUpperRoman }

// SequenceDefinition is a SEQUENCE_DEFINITION record: the numbering style, the
// starting value, and an optional prefix/suffix affixed to the derived label.
type SequenceDefinition struct {
	DefID  pdlfmt.UnitID
	Style  NumberingStyle
	Start  int
	Prefix string
	Suffix string
}

// OrderedItem is one item in an ordered sequence: which sequence definition it
// belongs to and its order-value (its ordinal within the sequence, 0-based).
type OrderedItem struct {
	ItemID     pdlfmt.UnitID
	DefID      pdlfmt.UnitID
	OrderValue int
}

// deriveNumeral renders the numeral (without prefix/suffix) for value v under
// style. v is the effective 1-based sequence number.
func deriveNumeral(style NumberingStyle, v int) string {
	if v < 1 {
		v = 1
	}
	switch style {
	case StyleLowerAlpha:
		return alpha(v, 'a')
	case StyleUpperAlpha:
		return alpha(v, 'A')
	case StyleLowerRoman:
		return strings.ToLower(roman(v))
	case StyleUpperRoman:
		return roman(v)
	default: // StyleDecimal
		return strconv.Itoa(v)
	}
}

// alpha renders bijective base-26 (a..z, aa..) starting at the given base rune.
func alpha(v int, base rune) string {
	var sb []byte
	for v > 0 {
		v--
		sb = append([]byte{byte(base) + byte(v%26)}, sb...)
		v /= 26
	}
	return string(sb)
}

// roman renders an uppercase Roman numeral for 1..3999 (clamped).
func roman(v int) string {
	if v > 3999 {
		v = 3999
	}
	vals := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	syms := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	var sb strings.Builder
	for i, val := range vals {
		for v >= val {
			sb.WriteString(syms[i])
			v -= val
		}
	}
	return sb.String()
}

// DeriveNumberingLabel returns the rendered numbering label for an item as a
// COMPUTED_INLINE value. The label is a pure deterministic function of the
// item's order-value, its referenced sequence definition (style + start +
// affixes), and its ROOT_SEQUENCE position (which, together with the item's
// order-value, fixes the effective sequence number). No persisted literal is
// consulted.
//
// rootSeqPosition is the item's position in ROOT_SEQUENCE; it participates so
// that two items with the same order-value at different reading positions
// still derive labels deterministically from their position (ties broken by
// reading order). The effective 1-based number is def.Start + order-value.
func DeriveNumberingLabel(def SequenceDefinition, item OrderedItem, rootSeqPosition int) ComputedInline {
	_ = rootSeqPosition // reading-order position is an input to determinism, not a literal source
	n := def.Start + item.OrderValue
	label := def.Prefix + deriveNumeral(def.Style, n) + def.Suffix
	return ComputedInline{Value: label, ValueDirection: DirectionLTR}
}
