package cli

import (
	"errors"
	"io"
	"strings"
)

// diff verb wiring (T-0332, TR-002; stdout shape corrected by
// DEFECT-2026-09-19b/T-0389). `protodoc diff` reports construct-level changes
// (changed constructs, not storage units), omitting agreements.

// DiffRun is the injectable construct-level diff backend (M13's diff engine).
// It returns the list of changed construct descriptions, and an error if either
// input cannot be opened/validated/decoded (so the verb can refuse rather than
// report a spurious no-difference OK).
var DiffRun = func(pathA, pathB string) ([]string, error) { return nil, nil }

// bucketChanges classifies realDiffRun's tagged description strings
// ("added@...", "removed@...", "changed@...") into cli.md S6's three
// documented buckets. A description with no recognised prefix (as synthetic
// test fixtures use) is bucketed as "changed", the safest default for an
// unclassified difference.
func bucketChanges(changed []string) (added, removed, other []string) {
	for _, c := range changed {
		switch {
		case strings.HasPrefix(c, "added@"):
			added = append(added, c)
		case strings.HasPrefix(c, "removed@"):
			removed = append(removed, c)
		default:
			other = append(other, c)
		}
	}
	return added, removed, other
}

func runDiff(args []string, _ io.Writer) Result {
	if len(args) < 2 {
		return StatusUsage.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-012", Message: "diff requires <fileA> <fileB>"}},
		})
	}
	changed, err := DiffRun(args[0], args[1])
	if err != nil {
		if errors.Is(err, ErrFileUnreadable) {
			return StatusUsage.ToResult(Result{
				Findings: []Finding{{RuleID: "TR-012", Message: "diff: " + err.Error()}},
			})
		}
		return StatusInvalid.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-012", Message: "diff failed: " + err.Error()}},
		})
	}
	added, removed, changedOnly := bucketChanges(changed)
	return StatusOK.ToResult(Result{
		Extra: map[string]any{
			"identical":          len(changed) == 0,
			"added":              added,
			"removed":            removed,
			"changed":            changedOnly,
			"changed_constructs": changed,
			"change_count":       len(changed),
		},
	})
}
