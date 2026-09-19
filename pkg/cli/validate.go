package cli

import (
	"io"

	"Protodoc/pkg/validate"
)

// validate verb wiring (T-0327, TR-012). `protodoc validate` runs the M07
// structural validation pipeline against a document and emits cli.md S2's
// checks/findings stdout shape, mapping the validity/budget verdicts through
// the exit-code engine (StatusInvalid / StatusOverBudget / StatusOK) via the
// existing validate->cli bridge.

// ValidateStepsFor builds the validation step set for a document path. It is a
// package variable so a test can inject at-limit / over-limit fixtures without
// a real file; the production wiring (a later CLI-integration task) replaces it
// with the container/ledger decode chain. It returns the steps and the number
// of checks run (for the "checks" envelope field).
var ValidateStepsFor = func(path string) ([]validate.Step, int) {
	return nil, 0
}

// runValidate executes the validate verb: it builds the document's steps, runs
// the pipeline, and returns a Result carrying the checks count, every finding
// (validity and budget), and the resolved status.
func runValidate(args []string, _ io.Writer) Result {
	if len(args) < 1 {
		return StatusUsage.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-012", Message: "validate requires a <file> argument"}},
		})
	}
	steps, checks := ValidateStepsFor(args[0])
	res := validate.Run(steps)

	// DEFECT-2026-09-19b/T-0390: a file that could not be opened at all is
	// USAGE, not INVALID (cli.md S1: "an unreadable path"), matching
	// inspect's already-correct behaviour.
	if res.Validity != nil && res.Validity.RuleID == unreadableFileRuleID {
		return StatusUsage.ToResult(Result{
			Findings: FindingsFromValidation(res),
			Extra:    map[string]any{"checks": checks},
		})
	}

	active := StatusesFromValidation(res)
	status := Resolve(active)
	return status.ToResult(Result{
		Findings: FindingsFromValidation(res),
		Extra:    map[string]any{"checks": checks},
	})
}
