package cli

import "io"

// redact verb wiring (T-0335, TR-012). `protodoc redact` removes designated
// subtrees via M11's redaction machinery; the published output declares the
// omissions (verify reports AttestedWithDeclaredOmissions) and contains none of
// the redacted plaintext.

// RedactResult is the redact backend's result.
type RedactResult struct {
	DeclaredOmissions []string // subtree ids declared as omitted
	Output            []byte   // the redacted output octets
	Err               error
}

// RedactRun is the injectable redaction backend (M11).
var RedactRun = func(path string, subtrees []string) RedactResult { return RedactResult{} }

func runRedact(args []string, _ io.Writer) Result {
	if len(args) < 1 {
		return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "redact requires a <file> argument"}}})
	}
	var subtrees []string
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--subtree" {
			subtrees = append(subtrees, args[i+1])
		}
	}
	out := RedactRun(args[0], subtrees)
	if out.Err != nil {
		return StatusInvalid.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: out.Err.Error()}}})
	}
	return StatusOK.ToResult(Result{
		Extra: map[string]any{"declared_omissions": out.DeclaredOmissions, "output_len": len(out.Output)},
	})
}
