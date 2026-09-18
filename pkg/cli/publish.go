package cli

import "io"

// publish verb wiring (T-0336, TR-012). `protodoc publish` re-emits a document
// via the M11/M17 canon.Publish path with zero octets of removed content and
// every custody/fixity value preserved unchanged.

// PublishResult is the publish backend's result.
type PublishResult struct {
	Output           []byte
	ResidueOctets    int  // count of removed-content octets found in output (must be 0)
	CustodyPreserved bool // every custody/fixity value matched the input unchanged
	Err              error
}

// PublishRun is the injectable publish backend (M11/M17).
var PublishRun = func(path string, partial bool) PublishResult { return PublishResult{CustodyPreserved: true} }

func runPublish(args []string, _ io.Writer) Result {
	if len(args) < 1 {
		return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "publish requires a <file> argument"}}})
	}
	partial := false
	for _, a := range args[1:] {
		if a == "--partial" {
			partial = true
		}
	}
	out := PublishRun(args[0], partial)
	if out.Err != nil {
		return StatusInvalid.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: out.Err.Error()}}})
	}
	if out.ResidueOctets != 0 || !out.CustodyPreserved {
		return StatusInvalid.ToResult(Result{Findings: []Finding{{RuleID: "FR-078", Message: "publish output failed residue/custody invariant"}}})
	}
	return StatusOK.ToResult(Result{
		Extra: map[string]any{"residue_octets": out.ResidueOctets, "custody_preserved": out.CustodyPreserved},
	})
}
