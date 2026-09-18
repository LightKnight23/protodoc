package semantics

import (
	"reflect"
	"testing"
)

// TestFR_034_ComputedInlineIsolation is T-0272's named conformance test
// (FR-034). Replacing a computed inline object's value with one of different
// directionality leaves the surrounding visual order identical, because the
// computed inline is isolated.
func TestFR_034_ComputedInlineIsolation(t *testing.T) {
	before := "page "
	after := " of 10"

	// Materialise with an LTR computed value.
	ltr := ComputedInline{Value: "7", ValueDirection: DirectionLTR}
	segsLTR := ResolveInlineSequence(before, ltr, after)
	orderLTR := SurroundingVisualOrder(segsLTR)

	// Re-materialise the SAME position with an RTL computed value (e.g. an
	// Arabic-Indic numeral string) of different directionality.
	rtl := ComputedInline{Value: "\u0663\u0664", ValueDirection: DirectionRTL} // ٣٤
	segsRTL := ResolveInlineSequence(before, rtl, after)
	orderRTL := SurroundingVisualOrder(segsRTL)

	// The surrounding visual order is identical -- the computed value's
	// directionality did not leak (FR-034).
	if !reflect.DeepEqual(orderLTR, orderRTL) {
		t.Errorf("surrounding order changed with the computed value: %v vs %v", orderLTR, orderRTL)
	}
	if !reflect.DeepEqual(orderLTR, []string{before, after}) {
		t.Errorf("surrounding order = %v, want %v", orderLTR, []string{before, after})
	}

	// The computed value is always the isolated segment.
	if len(segsLTR) != 3 || !segsLTR[1].Isolated {
		t.Errorf("computed inline must be the isolated middle segment: %+v", segsLTR)
	}
	// Surrounding segments are never isolated.
	if segsLTR[0].Isolated || segsLTR[2].Isolated {
		t.Error("surrounding text segments must not be isolated")
	}

	// Even a mixed-direction computed value does not perturb the surrounding
	// order.
	mixed := ComputedInline{Value: "A\u05D0B", ValueDirection: DirectionRTL} // AאB
	if got := SurroundingVisualOrder(ResolveInlineSequence(before, mixed, after)); !reflect.DeepEqual(got, orderLTR) {
		t.Errorf("mixed computed value perturbed surrounding order: %v", got)
	}
}
