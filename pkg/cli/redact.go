package cli

import (
	"errors"
	"io"
	"os"
)

// redact verb wiring (T-0335, TR-012; --subtree/--out enforcement and file-
// write fixed by DEFECT-2026-09-19b/T-0385). `protodoc redact` removes
// designated subtrees via M11's redaction machinery; the published output
// declares the omissions (verify reports AttestedWithDeclaredOmissions) and
// contains none of the redacted plaintext.

// RedactResult is the redact backend's result.
type RedactResult struct {
	DeclaredOmissions []string // subtree ids declared as omitted
	Output            []byte   // the redacted output octets
	Err               error
}

// RedactRun is the injectable redaction backend (M11).
var RedactRun = func(path string, subtrees []string) RedactResult { return RedactResult{} }

// writeRedactFile is injectable so tests can observe/stub the write.
var writeRedactFile = func(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}

func runRedact(args []string, _ io.Writer) Result {
	if len(args) < 1 {
		return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "redact requires a <file> argument"}}})
	}
	var subtrees []string
	to := ""
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--subtree":
			if i+1 < len(args) {
				subtrees = append(subtrees, args[i+1])
				i++
			}
		case "--out":
			if i+1 < len(args) {
				to = args[i+1]
				i++
			}
		}
	}
	if len(subtrees) == 0 {
		return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "redact requires at least one --subtree <unit-id>"}}})
	}
	if to == "" {
		return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "redact requires --out <path>"}}})
	}
	out := RedactRun(args[0], subtrees)
	if out.Err != nil {
		if errors.Is(out.Err, ErrFileUnreadable) {
			return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "redact: " + out.Err.Error()}}})
		}
		return StatusInvalid.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: out.Err.Error()}}})
	}
	if err := writeRedactFile(to, out.Output); err != nil {
		return StatusInvalid.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "redact: writing --out: " + err.Error()}}})
	}
	return StatusOK.ToResult(Result{
		Extra: map[string]any{"out": to, "declared_omissions": out.DeclaredOmissions, "output_len": len(out.Output)},
	})
}
