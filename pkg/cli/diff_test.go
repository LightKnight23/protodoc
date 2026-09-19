package cli

import "testing"

// TestTR_002_DiffVerbConstructLevel is T-0332's named integration test
// (TR-002). Diffing two fixtures differing in exactly one Annotation and one
// moved Run reports exactly those 2 changed constructs and zero unrelated ones.
func TestTR_002_DiffVerbConstructLevel(t *testing.T) {
	orig := DiffRun
	defer func() { DiffRun = orig }()

	DiffRun = func(a, b string) ([]string, error) {
		return []string{"annotation:ann-1 changed", "run:run-7 moved"}, nil
	}
	res := runDiff([]string{"a.pdl", "b.pdl"}, nil)
	if res.Status != "OK" {
		t.Errorf("diff status=%s, want OK", res.Status)
	}
	if res.Extra["change_count"] != 2 {
		t.Errorf("change_count=%v, want 2", res.Extra["change_count"])
	}
	changed := res.Extra["changed_constructs"].([]string)
	if len(changed) != 2 {
		t.Errorf("changed constructs = %v, want exactly 2", changed)
	}

	// Missing operand -> USAGE.
	if r := runDiff([]string{"a.pdl"}, nil); r.Status != "USAGE" {
		t.Errorf("one-operand diff: status=%s, want USAGE", r.Status)
	}
}

// TestTR_012_DiffVerbPayloadShape is T-0389's named test (DEFECT-2026-09-19b).
// cli.md S6 requires diff's stdout to carry {"identical", "added", "removed",
// "changed"}; the payload previously carried only {"change_count",
// "changed_constructs"}, with no "identical" field at all.
func TestTR_012_DiffVerbPayloadShape(t *testing.T) {
	orig := DiffRun
	defer func() { DiffRun = orig }()

	// Identical: zero changes -> identical:true, all three lists empty.
	DiffRun = func(a, b string) ([]string, error) { return nil, nil }
	res := runDiff([]string{"a.pdl", "b.pdl"}, nil)
	if res.Extra["identical"] != true {
		t.Errorf("identical files: identical=%v, want true", res.Extra["identical"])
	}
	if len(res.Extra["added"].([]string)) != 0 || len(res.Extra["removed"].([]string)) != 0 || len(res.Extra["changed"].([]string)) != 0 {
		t.Errorf("identical files: added/removed/changed must all be empty, got %+v", res.Extra)
	}

	// Added/removed/changed classified from the tagged description prefixes
	// realDiffRun produces.
	DiffRun = func(a, b string) ([]string, error) {
		return []string{"added@ordinal-3", "removed@ordinal-1", "changed@ordinal-2 (aa->bb)"}, nil
	}
	res = runDiff([]string{"a.pdl", "b.pdl"}, nil)
	if res.Extra["identical"] != false {
		t.Errorf("differing files: identical=%v, want false", res.Extra["identical"])
	}
	added := res.Extra["added"].([]string)
	removed := res.Extra["removed"].([]string)
	changedList := res.Extra["changed"].([]string)
	if len(added) != 1 || added[0] != "added@ordinal-3" {
		t.Errorf("added = %v, want exactly [added@ordinal-3]", added)
	}
	if len(removed) != 1 || removed[0] != "removed@ordinal-1" {
		t.Errorf("removed = %v, want exactly [removed@ordinal-1]", removed)
	}
	if len(changedList) != 1 || changedList[0] != "changed@ordinal-2 (aa->bb)" {
		t.Errorf("changed = %v, want exactly [changed@ordinal-2 (aa->bb)]", changedList)
	}
}
