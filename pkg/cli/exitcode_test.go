package cli

import (
	"bytes"
	"testing"
)

// TestTR_012_ExitCodePrecedenceTable is T-0326's named test. It enumerates
// every pairwise combination of cli.md S1's 8 documented statuses and
// confirms Resolve always returns the single precedence winner named by
// S1.1's ranking, with no ambiguous or undocumented pair.
func TestTR_012_ExitCodePrecedenceTable(t *testing.T) {
	if len(precedenceOrder) != len(allStatuses) {
		t.Fatalf("precedenceOrder has %d entries, want %d (one per documented status)", len(precedenceOrder), len(allStatuses))
	}

	rank := make(map[Status]int, len(precedenceOrder))
	for i, s := range precedenceOrder {
		if _, dup := rank[s]; dup {
			t.Fatalf("status %s appears more than once in precedenceOrder", s)
		}
		rank[s] = i
	}
	for _, s := range allStatuses {
		if _, ok := rank[s]; !ok {
			t.Fatalf("status %s is documented but missing from precedenceOrder", s)
		}
	}

	for _, a := range allStatuses {
		for _, b := range allStatuses {
			active := map[Status]bool{a: true, b: true}
			got := Resolve(active)

			want := a
			if rank[b] < rank[a] {
				want = b
			}
			if got != want {
				t.Errorf("Resolve({%s, %s}) = %s, want %s (the higher-ranked of the pair)", a, b, got, want)
			}
			if got != a && got != b {
				t.Errorf("Resolve({%s, %s}) = %s: winner is neither input, an undocumented pair", a, b, got)
			}
		}
	}
}

// TestTR_012_ExitCodePrecedenceSingletons confirms every status alone
// resolves to itself, and that an empty condition set is OK.
func TestTR_012_ExitCodePrecedenceSingletons(t *testing.T) {
	for _, s := range allStatuses {
		if got := Resolve(map[Status]bool{s: true}); got != s {
			t.Errorf("Resolve({%s}) = %s, want %s", s, got, s)
		}
	}
	if got := Resolve(map[Status]bool{}); got != StatusOK {
		t.Errorf("Resolve({}) = %s, want OK", got)
	}
}

// TestTR_012_ExitCodeTableValues locks cli.md S1's exit-code integers and
// status name spellings so a future edit cannot silently renumber them.
func TestTR_012_ExitCodeTableValues(t *testing.T) {
	want := map[Status]struct {
		code int
		name string
	}{
		StatusOK:          {0, "OK"},
		StatusInvalid:     {1, "INVALID"},
		StatusUnsupported: {2, "UNSUPPORTED"},
		StatusUnverified:  {3, "UNVERIFIED"},
		StatusUnavailable: {4, "UNAVAILABLE"},
		StatusOverBudget:  {5, "OVER_BUDGET"},
		StatusRefused:     {6, "REFUSED"},
		StatusUsage:       {7, "USAGE"},
	}
	if len(want) != len(allStatuses) {
		t.Fatalf("test fixture lists %d statuses, want %d", len(want), len(allStatuses))
	}
	for s, w := range want {
		if s.Code() != w.code {
			t.Errorf("%s.Code() = %d, want %d", s, s.Code(), w.code)
		}
		if s.String() != w.name {
			t.Errorf("%s.String() = %q, want %q", s, s.String(), w.name)
		}
	}
}

// TestTR_012_ToResultSetsStatusAndExitCode confirms Status.ToResult is the
// single place a verb's RunFunc converts a resolved Status into the
// envelope's status/exit_code fields, preserving Findings/Extra untouched.
func TestTR_012_ToResultSetsStatusAndExitCode(t *testing.T) {
	base := Result{
		Findings: []Finding{{RuleID: "FR-104", Message: "example"}},
		Extra:    map[string]any{"checks": map[string]string{"header_and_capability": "OK"}},
	}
	got := StatusInvalid.ToResult(base)

	if got.Status != "INVALID" {
		t.Errorf("Status = %q, want INVALID", got.Status)
	}
	if got.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", got.ExitCode)
	}
	if len(got.Findings) != 1 || got.Findings[0].RuleID != "FR-104" {
		t.Errorf("Findings not preserved: %+v", got.Findings)
	}
	if got.Extra["checks"] == nil {
		t.Error("Extra not preserved")
	}
}

// TestTR_012_NotImplementedUsesExitCodeEngine confirms the dispatch
// framework's own placeholder verb result is produced through the shared
// exit-code engine (StatusUsage.ToResult), not a hand-typed duplicate of
// its status name and exit code.
func TestTR_012_NotImplementedUsesExitCodeEngine(t *testing.T) {
	var stderr bytes.Buffer
	result := notImplemented("validate")(nil, &stderr)
	if result.Status != StatusUsage.String() {
		t.Errorf("Status = %q, want %q", result.Status, StatusUsage.String())
	}
	if result.ExitCode != StatusUsage.Code() {
		t.Errorf("ExitCode = %d, want %d", result.ExitCode, StatusUsage.Code())
	}
}
