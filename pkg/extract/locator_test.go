package extract_test

import (
	"testing"
	"unicode/utf8"

	"Protodoc/pkg/content"
	"Protodoc/pkg/extract"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_042_EmitsUnitIdentityAndScalarPositionLocator is T-0088's named
// test. For several multi-run text units (one containing combining-mark
// clusters), it asserts each emitted locator's UnitID matches the run's
// identity and its ScalarOffset matches an independently-computed scalar
// count -- confirming the offset is live scalar counting scoped to the unit
// (CON-001), not a persisted position.
func TestFR_042_EmitsUnitIdentityAndScalarPositionLocator(t *testing.T) {
	mkRun := func(seed byte, text string) content.Run {
		var id pdlfmt.UnitID
		for i := range id {
			id[i] = seed + byte(i)
		}
		return content.Run{RunID: id, Text: text, LangRef: content.LangRef(1)}
	}

	fixtures := [][]content.Run{
		// Plain ASCII multi-run.
		{mkRun(1, "hello"), mkRun(20, " world"), mkRun(40, "!")},
		// Astral + BMP mix.
		{mkRun(60, "a\U0001D11E"), mkRun(80, "\u4e2db")},
		// Combining-mark clusters: each combining mark is its own scalar.
		{mkRun(100, "e\u0301"), mkRun(120, "o\u0308o\u0308"), mkRun(140, "z")},
	}

	for fi, runs := range fixtures {
		locs := extract.LocatorsForUnit(runs)
		if len(locs) != len(runs) {
			t.Fatalf("fixture %d: emitted %d locators, want %d", fi, len(locs), len(runs))
		}
		var want uint32
		for i, r := range runs {
			// UnitID matches the run identity.
			if locs[i].UnitID != r.RunID {
				t.Fatalf("fixture %d run %d: locator UnitID mismatch", fi, i)
			}
			// ScalarOffset matches the independently-counted running scalar
			// total of preceding runs (utf8.RuneCountInString is the
			// independent oracle for scalar-value count).
			if locs[i].ScalarOffset != want {
				t.Fatalf("fixture %d run %d: ScalarOffset = %d, want %d", fi, i, locs[i].ScalarOffset, want)
			}
			want += uint32(utf8.RuneCountInString(r.Text))
		}
	}

	// Point addressing: LocatorAt returns the run identity and the given
	// scalar index; out-of-range is rejected.
	r := mkRun(200, "a\u0301bc") // 4 scalars
	loc, ok := extract.LocatorAt(r, 2)
	if !ok || loc.UnitID != r.RunID || loc.ScalarOffset != 2 {
		t.Fatalf("LocatorAt(r,2) = %+v ok=%v, want {rid,2} true", loc, ok)
	}
	if _, ok := extract.LocatorAt(r, 4); ok {
		t.Fatalf("LocatorAt at length unexpectedly ok")
	}
}
