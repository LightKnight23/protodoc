package cli

import "testing"

// TestTR_003_MergeVerbConflictExit is T-0333's named integration test (TR-003).
// A genuine R2/R3 conflict exits CONFLICT naming both source values; a
// CON-024 retention-point-crossing and a CON-025 history-mode-mismatch each
// exit REFUSED naming the specific violated condition.
func TestTR_003_MergeVerbConflictExit(t *testing.T) {
	orig := MergeRun
	defer func() { MergeRun = orig }()

	// R2/R3 conflict -> CONFLICT result, INVALID exit, both values named.
	MergeRun = func(base, a, b string) MergeOutcome {
		return MergeOutcome{Kind: MergeConflict, ValueA: "LTR", ValueB: "RTL"}
	}
	res := runMerge([]string{"base", "a", "b"}, nil)
	if res.Extra["result"] != "CONFLICT" || res.ExitCode != StatusInvalid.Code() {
		t.Errorf("conflict: result=%v exit=%d, want CONFLICT/%d", res.Extra["result"], res.ExitCode, StatusInvalid.Code())
	}
	if res.Extra["value_a"] != "LTR" || res.Extra["value_b"] != "RTL" {
		t.Errorf("conflict must name both source values, got %v / %v", res.Extra["value_a"], res.Extra["value_b"])
	}

	// CON-024 retention-point crossing -> REFUSED naming the condition.
	MergeRun = func(base, a, b string) MergeOutcome {
		return MergeOutcome{Kind: MergeRefused, RefusedCondition: "CON-024"}
	}
	res = runMerge([]string{"base", "a", "b"}, nil)
	if res.Status != "REFUSED" || res.Extra["condition"] != "CON-024" {
		t.Errorf("retention crossing: status=%s condition=%v, want REFUSED/CON-024", res.Status, res.Extra["condition"])
	}

	// CON-025 history-mode mismatch -> REFUSED naming the condition.
	MergeRun = func(base, a, b string) MergeOutcome {
		return MergeOutcome{Kind: MergeRefused, RefusedCondition: "CON-025"}
	}
	res = runMerge([]string{"base", "a", "b"}, nil)
	if res.Status != "REFUSED" || res.Extra["condition"] != "CON-025" {
		t.Errorf("history-mode mismatch: status=%s condition=%v, want REFUSED/CON-025", res.Status, res.Extra["condition"])
	}

	// Clean merge -> OK.
	MergeRun = func(base, a, b string) MergeOutcome { return MergeOutcome{Kind: MergeClean} }
	if r := runMerge([]string{"base", "a", "b"}, nil); r.Status != "OK" {
		t.Errorf("clean merge: status=%s, want OK", r.Status)
	}

	// Missing operands -> USAGE.
	if r := runMerge([]string{"base", "a"}, nil); r.Status != "USAGE" {
		t.Errorf("two-operand merge: status=%s, want USAGE", r.Status)
	}
}
