package cli

import "io"

// diff verb wiring (T-0332, TR-002). `protodoc diff` reports construct-level
// changes (changed constructs, not storage units), omitting agreements.

// DiffRun is the injectable construct-level diff backend (M13's diff engine).
// It returns the list of changed construct descriptions.
var DiffRun = func(pathA, pathB string) []string { return nil }

func runDiff(args []string, _ io.Writer) Result {
	if len(args) < 2 {
		return StatusUsage.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-012", Message: "diff requires <fileA> <fileB>"}},
		})
	}
	changed := DiffRun(args[0], args[1])
	return StatusOK.ToResult(Result{
		Extra: map[string]any{
			"changed_constructs": changed,
			"change_count":       len(changed),
		},
	})
}
