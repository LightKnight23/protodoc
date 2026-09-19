package cli

import (
	"errors"
	"io"
	"os"
)

// publish verb wiring (T-0336, TR-012; --out enforcement and file-write fixed
// by DEFECT-2026-09-19b/T-0386). `protodoc publish` re-emits a document via
// the M11/M17 canon.Publish path with zero octets of removed content and
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

// writePublishFile is injectable so tests can observe/stub the write.
var writePublishFile = func(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}

func runPublish(args []string, _ io.Writer) Result {
	if len(args) < 1 {
		return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "publish requires a <file> argument"}}})
	}
	partial := false
	to := ""
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--partial":
			partial = true
		case "--out":
			if i+1 < len(args) {
				to = args[i+1]
				i++
			}
		}
	}
	if to == "" {
		return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "publish requires --out <path>"}}})
	}
	out := PublishRun(args[0], partial)
	if out.Err != nil {
		if errors.Is(out.Err, ErrFileUnreadable) {
			return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "publish: " + out.Err.Error()}}})
		}
		return StatusInvalid.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: out.Err.Error()}}})
	}
	if out.ResidueOctets != 0 || !out.CustodyPreserved {
		return StatusInvalid.ToResult(Result{Findings: []Finding{{RuleID: "FR-078", Message: "publish output failed residue/custody invariant"}}})
	}
	if err := writePublishFile(to, out.Output); err != nil {
		return StatusInvalid.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "publish: writing --out: " + err.Error()}}})
	}
	return StatusOK.ToResult(Result{
		Extra: map[string]any{"out": to, "residue_octets": out.ResidueOctets, "custody_preserved": out.CustodyPreserved},
	})
}
