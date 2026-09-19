package cli

import (
	"errors"
	"io"
	"os"

	"Protodoc/pkg/canon"
)

// project verb wiring (T-0334, TR-004; --to enforcement and file-write fixed
// by DEFECT-2026-09-19b/T-0383). `protodoc project` emits M17 canon's
// deterministic text/html projection of a document state (--format=text|html)
// and writes it to the required --to path. The projection is a pure function
// of the logical state (canon.ProjectText / canon.ProjectHTML).

// ProjectStateFor loads a document path into a canon.Document state. It is
// injectable so tests can drive projections without a real file; production
// wiring replaces it with the container/ledger decode chain.
var ProjectStateFor = func(path string) (*canon.Document, error) { return &canon.Document{}, nil }

// writeProjectionFile is injectable so tests can observe/stub the write
// without touching a real filesystem path.
var writeProjectionFile = func(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}

func runProject(args []string, _ io.Writer) Result {
	if len(args) < 1 {
		return StatusUsage.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-012", Message: "project requires a <file> argument"}},
		})
	}
	format := ""
	to := ""
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--format=html":
			format = "html"
		case "--format=text":
			format = "text"
		case "--to":
			if i+1 < len(args) {
				to = args[i+1]
				i++
			}
		}
	}
	if format == "" {
		format = "text"
	}
	if to == "" {
		return StatusUsage.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-012", Message: "project requires --to <path>"}},
		})
	}
	state, err := ProjectStateFor(args[0])
	if err != nil {
		if errors.Is(err, ErrFileUnreadable) {
			return StatusUsage.ToResult(Result{
				Findings: []Finding{{RuleID: "TR-012", Message: "project: " + err.Error()}},
			})
		}
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
	if err := writeProjectionFile(to, []byte(projection)); err != nil {
		return StatusInvalid.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-012", Message: "project: writing --to: " + err.Error()}},
		})
	}
	return StatusOK.ToResult(Result{
		Extra: map[string]any{"out": to, "format": format, "reingestable": false},
	})
}
