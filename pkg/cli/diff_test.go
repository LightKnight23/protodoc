package cli

import "testing"

// TestTR_002_DiffVerbConstructLevel is T-0332's named integration test
// (TR-002). Diffing two fixtures differing in exactly one Annotation and one
// moved Run reports exactly those 2 changed constructs and zero unrelated ones.
func TestTR_002_DiffVerbConstructLevel(t *testing.T) {
	orig := DiffRun
	defer func() { DiffRun = orig }()

	DiffRun = func(a, b string) []string {
		return []string{"annotation:ann-1 changed", "run:run-7 moved"}
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
