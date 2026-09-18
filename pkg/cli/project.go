package cli

import (
	"io"

	"Protodoc/pkg/canon"
)

// project verb wiring (T-0334, TR-004). `protodoc project` emits M17 canon's
// deterministic text/html projection of a document state (--format=text|html).
// The projection is a pure function of the logical state (canon.ProjectText /
// canon.ProjectHTML).

// ProjectStateFor loads a document path into a canon.Document state. It is
// injectable so tests can drive projections without a real file; production
// wiring replaces it with the container/ledger decode chain.
var ProjectStateFor = func(path string) (*canon.Document, error) { return &canon.Document{}, nil }

func runProject(args []string, _ io.Writer) Result {
	if len(args) < 1 {
		return StatusUsage.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-012", Message: "project requires a <file> argument"}},
		})
	}
	format := "text"
	for _, a := range args[1:] {
		switch a {
		case "--format=html":
			format = "html"
		case "--format=text":
			format = "text"
		}
	}
	state, err := ProjectStateFor(args[0])
	if err != nil {
		return StatusInvalid.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-012", Message: "project failed: " + err.Error()}},
		})
	}
	var projection string
	switch format {
	case "html":
		projection = canon.ProjectHTML(state)
	default:
		projection = canon.ProjectText(state)
	}
	return StatusOK.ToResult(Result{
		Extra: map[string]any{"format": format, "projection": projection},
	})
}
