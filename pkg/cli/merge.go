package cli

import "io"

// merge verb wiring (T-0333, TR-003). `protodoc merge` runs M13's DP-015 merge
// classifier, surfacing a genuine R2/R3 conflict as CONFLICT (naming both
// source values) and a precondition violation (CON-024 retention-point crossing,
// CON-025 history-mode mismatch) as REFUSED naming the specific condition.

// MergeOutcomeKind classifies a merge result.
type MergeOutcomeKind int

const (
	MergeClean MergeOutcomeKind = iota
	MergeConflict
	MergeRefused
)

// MergeOutcome is the merge backend's result.
type MergeOutcome struct {
	Kind MergeOutcomeKind
	// Conflict values (for MergeConflict): both source values named.
	ValueA, ValueB string
	// Refusal condition (for MergeRefused), e.g. "CON-024" / "CON-025".
	RefusedCondition string
}

// MergeRun is the injectable merge classifier backend (M13 DP-015).
var MergeRun = func(base, a, b string) MergeOutcome { return MergeOutcome{} }

// runMerge classifies a three-way merge: a genuine R2/R3 conflict is reported
// as a CONFLICT result (INVALID exit — no merged output is produced, naming
// both source values); a precondition violation is REFUSED naming the specific
// condition (CON-024/CON-025); otherwise the merge is clean (OK).
func runMerge(args []string, _ io.Writer) Result {
	if len(args) < 3 {
		return StatusUsage.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-012", Message: "merge requires <base> <a> <b>"}},
		})
	}
	out := MergeRun(args[0], args[1], args[2])
	switch out.Kind {
	case MergeConflict:
		return StatusInvalid.ToResult(Result{
			Extra: map[string]any{
				"result":  "CONFLICT",
				"value_a": out.ValueA,
				"value_b": out.ValueB,
			},
			Findings: []Finding{{RuleID: "TR-003", Message: "merge conflict: " + out.ValueA + " vs " + out.ValueB}},
		})
	case MergeRefused:
		return StatusRefused.ToResult(Result{
			Extra:    map[string]any{"result": "REFUSED", "condition": out.RefusedCondition},
			Findings: []Finding{{RuleID: out.RefusedCondition, Message: "merge refused: " + out.RefusedCondition}},
		})
	default:
		return StatusOK.ToResult(Result{Extra: map[string]any{"result": "MERGED"}})
	}
}
