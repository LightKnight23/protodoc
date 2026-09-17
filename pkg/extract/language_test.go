package extract_test

import (
	"errors"
	"testing"

	"Protodoc/pkg/content"
	"Protodoc/pkg/extract"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_043_EmitsExactlyOneLanguageTagPerUnit is T-0090's named conformance
// test. Every unit in a 3-language mixed fixture reports exactly one
// resolved tag; a unit whose language reference is unresolvable is reported
// via a per-unit error value in the result, never silently omitted.
func TestFR_043_EmitsExactlyOneLanguageTagPerUnit(t *testing.T) {
	mkRun := func(seed byte, lang content.LangRef) content.Run {
		var id pdlfmt.UnitID
		id[0] = seed
		return content.Run{RunID: id, Text: "x", LangRef: lang}
	}

	reg := content.MapLanguageRegistry{
		content.LangRef(1): "en-US",
		content.LangRef(2): "ja-JP",
		content.LangRef(3): "ar-EG",
	}

	// Three units in three languages, plus one with an unresolvable ref.
	runs := []content.Run{
		mkRun(1, content.LangRef(1)),
		mkRun(2, content.LangRef(2)),
		mkRun(3, content.LangRef(3)),
		mkRun(4, content.LangRef(99)), // unresolvable
	}

	results := extract.LanguagesForRuns(runs, reg)

	// No unit omitted: one result per run.
	if len(results) != len(runs) {
		t.Fatalf("got %d language results, want one per unit = %d", len(results), len(runs))
	}

	wantTags := []string{"en-US", "ja-JP", "ar-EG"}
	for i := 0; i < 3; i++ {
		if results[i].Err != nil {
			t.Fatalf("unit %d unexpectedly errored: %v", i, results[i].Err)
		}
		if results[i].Tag != wantTags[i] {
			t.Fatalf("unit %d tag = %q, want %q (exactly one resolved tag)", i, results[i].Tag, wantTags[i])
		}
		if results[i].UnitID != runs[i].RunID {
			t.Fatalf("unit %d identity mismatch", i)
		}
	}

	// The unresolvable unit is present with a per-unit error and no tag.
	last := results[3]
	if last.Err == nil {
		t.Fatalf("unresolvable unit did not report an error (must not be silently omitted)")
	}
	if !errors.Is(last.Err, content.ErrUnresolvableLanguageTag) {
		t.Fatalf("unresolvable unit error = %v, want ErrUnresolvableLanguageTag", last.Err)
	}
	if last.Tag != "" {
		t.Fatalf("unresolvable unit reported a tag %q, want empty", last.Tag)
	}

	// An unset reference also produces a per-unit error, not omission.
	unset := extract.LanguageOf(runs[0].RunID, content.LangUnset, reg)
	if unset.Err == nil || !errors.Is(unset.Err, content.ErrMissingLanguageRef) {
		t.Fatalf("unset reference error = %v, want ErrMissingLanguageRef", unset.Err)
	}
}
