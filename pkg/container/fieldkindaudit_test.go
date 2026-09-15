package container

import "testing"

// TestCON_007_NoFieldResolvesLocationOrExecutes is T-0024's named test.
// Implements: CON-007.
//
// It walks FieldKindVocabulary — the single closed list covering both the
// PDL-TLV field-kind vocabulary (pkg/pdlfmt, T-0002) and the container's
// fixed-prefix field set — and asserts every entry's category is one of
// the closed, non-location, non-executable categories, and none is one of
// the categories CON-007 forbids. Extending FieldKindVocabulary with a new
// field-kind re-runs this same audit automatically: no change to this
// test is needed.
func TestCON_007_NoFieldResolvesLocationOrExecutes(t *testing.T) {
	if len(FieldKindVocabulary) == 0 {
		t.Fatal("FieldKindVocabulary is empty, audit has nothing to check")
	}

	// closedFieldKindCategories and forbiddenFieldKindCategories must
	// never overlap: if they did, a category could pass the "is it in
	// the closed set" check while also being one CON-007 forbids.
	for cat := range closedFieldKindCategories {
		if forbiddenFieldKindCategories[cat] {
			t.Fatalf("category %q is in both the closed set and the forbidden set", cat)
		}
	}

	seen := make(map[string]bool, len(FieldKindVocabulary))
	for _, fk := range FieldKindVocabulary {
		if fk.Name == "" {
			t.Fatalf("field-kind with empty Name (category %q)", fk.Category)
		}
		if seen[fk.Name] {
			t.Fatalf("field-kind %q listed more than once in FieldKindVocabulary", fk.Name)
		}
		seen[fk.Name] = true

		if forbiddenFieldKindCategories[fk.Category] {
			t.Fatalf("field-kind %q has category %q, a network/filesystem-location or executable-valued category forbidden by CON-007", fk.Name, fk.Category)
		}
		if !closedFieldKindCategories[fk.Category] {
			t.Fatalf("field-kind %q has category %q, outside the closed CON-007 category set entirely", fk.Name, fk.Category)
		}
	}
}
