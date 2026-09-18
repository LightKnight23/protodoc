package semantics

import "testing"

// TestFR_083_NumberingLabelDeterministic is T-0286's named unit test (FR-083).
// A rendered numbering label is a pure deterministic function of (order-value,
// sequence definition, ROOT_SEQUENCE position), carried as a COMPUTED_INLINE
// value; no persisted literal is authoritative.
func TestFR_083_NumberingLabelDeterministic(t *testing.T) {
	def := SequenceDefinition{DefID: pdUnitSem(0x01), Style: StyleDecimal, Start: 1, Suffix: "."}

	// Determinism: identical inputs give identical labels across calls.
	item := OrderedItem{ItemID: pdUnitSem(0x10), DefID: def.DefID, OrderValue: 2}
	a := DeriveNumberingLabel(def, item, 5)
	b := DeriveNumberingLabel(def, item, 5)
	if a != b {
		t.Errorf("label not deterministic: %+v vs %+v", a, b)
	}
	if a.Value != "3." { // start 1 + order 2 = 3
		t.Errorf("decimal label = %q, want %q", a.Value, "3.")
	}

	// It is a COMPUTED_INLINE (isolated when materialised into surrounding text).
	segs := ResolveInlineSequence("see item ", a, " above")
	if len(segs) != 3 || !segs[1].Isolated || segs[1].Text != "3." {
		t.Errorf("numbering label must materialise as an isolated computed inline, got %+v", segs)
	}

	// Style coverage: each style derives the expected numeral for order 0..3.
	styleCases := []struct {
		style NumberingStyle
		want  []string
	}{
		{StyleDecimal, []string{"1", "2", "3", "4"}},
		{StyleLowerAlpha, []string{"a", "b", "c", "d"}},
		{StyleUpperAlpha, []string{"A", "B", "C", "D"}},
		{StyleLowerRoman, []string{"i", "ii", "iii", "iv"}},
		{StyleUpperRoman, []string{"I", "II", "III", "IV"}},
	}
	for _, sc := range styleCases {
		d := SequenceDefinition{Style: sc.style, Start: 1}
		for ov := 0; ov < 4; ov++ {
			got := DeriveNumberingLabel(d, OrderedItem{OrderValue: ov}, ov)
			if got.Value != sc.want[ov] {
				t.Errorf("style %d order %d: label = %q, want %q", sc.style, ov, got.Value, sc.want[ov])
			}
		}
	}

	// Bijective base-26 rollover: order 26 (n=27) -> "aa".
	dAlpha := SequenceDefinition{Style: StyleLowerAlpha, Start: 1}
	if got := DeriveNumberingLabel(dAlpha, OrderedItem{OrderValue: 26}, 0); got.Value != "aa" {
		t.Errorf("lower-alpha n=27 = %q, want %q", got.Value, "aa")
	}
}
