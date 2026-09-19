package cli

import (
	"path/filepath"
	"testing"
)

// TestTR_012_DiffVerbReadsRealFiles is T-0377's named integration test
// (TR-012/TR-002, DEFECT-2026-09-19 fix). The production diff backend
// (realDiffRun, wired by init) reads two real files: identical documents report
// zero changed constructs; documents differing in a CONTENT segment report that
// construct as changed; an absent input yields a diff-error sentinel, never a
// silent no-difference.
func TestTR_012_DiffVerbReadsRealFiles(t *testing.T) {
	// (1) Two identical real documents -> zero changed constructs.
	a := writeDocWithContentSegments(t, 2, 256)
	b := writeDocWithContentSegments(t, 2, 256) // same construction => identical slot digests
	res := runDiff([]string{a, b}, nil)
	if res.Status != "OK" {
		t.Fatalf("identical docs: status=%s, want OK", res.Status)
	}
	if res.Extra["change_count"].(int) != 0 {
		t.Errorf("identical documents should report 0 changes, got %d", res.Extra["change_count"])
	}

	// (2) Documents with a differing content-segment count -> changes reported.
	c := writeDocWithContentSegments(t, 3, 256)
	res = runDiff([]string{a, c}, nil)
	if res.Extra["change_count"].(int) == 0 {
		t.Errorf("documents differing in content must report changes, got 0")
	}

	// (3) Absent input -> INVALID (the verb refuses rather than reporting a
	// spurious no-difference OK).
	res = runDiff([]string{a, filepath.Join(t.TempDir(), "absent.pdl")}, nil)
	if res.Status != "INVALID" {
		t.Errorf("absent input must yield INVALID, got %s", res.Status)
	}

	// Missing operand -> USAGE.
	if r := runDiff([]string{a}, nil); r.Status != "USAGE" {
		t.Errorf("one-operand diff: status=%s, want USAGE", r.Status)
	}
}
