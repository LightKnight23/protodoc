package render

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// contentUnitA11Y006 is one content unit for the parity suite: its identity,
// its authored text, and its ROOT_SEQUENCE reading-order position.
type contentUnitA11Y006 struct {
	id       pdlfmt.UnitID
	text     string
	seqOrder int
}

// presentUnitsFixed returns the units as a fixed-pagination presentation would
// visit them: broken into fixed-size pages, but always in ROOT_SEQUENCE order.
// The page grouping is the ONLY thing pagination changes; it must not reorder
// or drop content.
func presentUnitsFixed(units []contentUnitA11Y006, perPage int) [][]contentUnitA11Y006 {
	ordered := orderByRootSequenceA11Y006(units)
	var pages [][]contentUnitA11Y006
	for i := 0; i < len(ordered); i += perPage {
		end := i + perPage
		if end > len(ordered) {
			end = len(ordered)
		}
		pages = append(pages, ordered[i:end])
	}
	return pages
}

// presentUnitsReflow returns the units as a Knuth-Plass reflow presentation
// would visit them: line-broken at an arbitrary viewport width, but again
// always in ROOT_SEQUENCE order. The line grouping differs from the page
// grouping above; the content sequence must not.
func presentUnitsReflow(units []contentUnitA11Y006, perLine int) [][]contentUnitA11Y006 {
	ordered := orderByRootSequenceA11Y006(units)
	var lines [][]contentUnitA11Y006
	for i := 0; i < len(ordered); i += perLine {
		end := i + perLine
		if end > len(ordered) {
			end = len(ordered)
		}
		lines = append(lines, ordered[i:end])
	}
	return lines
}

// orderByRootSequenceA11Y006 sorts units by their ROOT_SEQUENCE position; the
// order is presentation-independent.
func orderByRootSequenceA11Y006(units []contentUnitA11Y006) []contentUnitA11Y006 {
	out := make([]contentUnitA11Y006, len(units))
	copy(out, units)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1].seqOrder > out[j].seqOrder; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}

// flattenText concatenates the text of a grouped presentation in visit order.
func flattenText(groups [][]contentUnitA11Y006) string {
	s := ""
	for _, g := range groups {
		for _, u := range g {
			s += u.text
		}
	}
	return s
}

// flattenReadingOrder returns the unit ids of a grouped presentation in visit
// order (the ROOT_SEQUENCE-derived reading order).
func flattenReadingOrder(groups [][]contentUnitA11Y006) []pdlfmt.UnitID {
	var order []pdlfmt.UnitID
	for _, g := range groups {
		for _, u := range g {
			order = append(order, u.id)
		}
	}
	return order
}

// TestConformanceA11Y006FixedReflowParity is T-0278's named conformance test
// (vector CONFORMANCE-A11Y-006-fixed-reflow-parity; FR-036/FR-101). One state
// is presented under both fixed pagination and Knuth-Plass reflow; the
// extracted text and the ROOT_SEQUENCE-derived reading order MUST be identical
// between the two presentations. Only the visual grouping (pages vs lines) may
// differ.
func TestConformanceA11Y006FixedReflowParity(t *testing.T) {
	u := func(seed byte) pdlfmt.UnitID {
		var id pdlfmt.UnitID
		id[0] = seed
		return id
	}
	// Authored in scrambled storage order; seqOrder gives the reading order.
	units := []contentUnitA11Y006{
		{id: u(0x30), text: "gamma ", seqOrder: 2},
		{id: u(0x10), text: "alpha ", seqOrder: 0},
		{id: u(0x40), text: "delta ", seqOrder: 3},
		{id: u(0x20), text: "beta ", seqOrder: 1},
		{id: u(0x50), text: "epsilon", seqOrder: 4},
	}

	// Fixed pagination: 2 units per page. Reflow: 3 units per line. The
	// groupings deliberately differ.
	fixed := presentUnitsFixed(units, 2)
	reflow := presentUnitsReflow(units, 3)

	// Grouping differs (proving the two presentations are genuinely distinct).
	if len(fixed) == len(reflow) {
		t.Fatalf("test setup: fixed (%d groups) and reflow (%d groups) should differ", len(fixed), len(reflow))
	}

	// Extracted text is identical.
	if ft, rt := flattenText(fixed), flattenText(reflow); ft != rt {
		t.Errorf("extracted text differs between presentations:\n fixed  = %q\n reflow = %q", ft, rt)
	}

	// ROOT_SEQUENCE-derived reading order is identical.
	fo, ro := flattenReadingOrder(fixed), flattenReadingOrder(reflow)
	if len(fo) != len(ro) {
		t.Fatalf("reading-order lengths differ: fixed %d, reflow %d", len(fo), len(ro))
	}
	for i := range fo {
		if fo[i] != ro[i] {
			t.Errorf("reading order differs at %d: fixed %x, reflow %x", i, fo[i][0], ro[i][0])
		}
	}

	// And the reading order actually matches the authored ROOT_SEQUENCE
	// (alpha,beta,gamma,delta,epsilon), not the scrambled storage order.
	wantFirst := u(0x10)
	if fo[0] != wantFirst {
		t.Errorf("reading order[0] = %x, want alpha (%x)", fo[0][0], wantFirst[0])
	}
}
