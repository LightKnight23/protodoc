package merge

import (
	"bytes"
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestTR_003_ConflictNamesBothValuesNoAutoResolve is T-0240's named unit test
// (TR-003). When a three-way merge hits a conflict it names BOTH values,
// produces no merged value for that construct, and never selects a side,
// concatenates, or interleaves content. The CLI turns a non-empty conflict set
// into a non-zero exit.
func TestTR_003_ConflictNamesBothValuesNoAutoResolve(t *testing.T) {
	conflicted := mTarget(0x01)
	leftOnly := mTarget(0x02)
	agreed := mTarget(0x03)

	base := map[pdlfmt.UnitID][]byte{
		conflicted: []byte("base"),
		leftOnly:   []byte("keep"),
		agreed:     []byte("same-old"),
	}
	left := map[pdlfmt.UnitID][]byte{
		conflicted: []byte("LEFT-VALUE"),
		leftOnly:   []byte("left-edit"),
		agreed:     []byte("same-new"),
	}
	right := map[pdlfmt.UnitID][]byte{
		conflicted: []byte("RIGHT-VALUE"),
		leftOnly:   []byte("keep"),     // unchanged on right
		agreed:     []byte("same-new"), // same change as left
	}

	res := ThreeWayMerge(base, left, right)

	// A conflict must be reported (non-zero exit signal).
	if !res.HasConflict() {
		t.Fatal("divergent edit on both sides must produce a conflict")
	}

	// Exactly the conflicted construct is unresolved, naming both values.
	if len(res.Conflicts) != 1 {
		t.Fatalf("got %d conflicts, want 1: %+v", len(res.Conflicts), res.Conflicts)
	}
	c := res.Conflicts[0]
	if c.Construct != conflicted {
		t.Errorf("conflict names construct %x, want %x", c.Construct, conflicted)
	}
	if !bytes.Equal(c.Left, []byte("LEFT-VALUE")) || !bytes.Equal(c.Right, []byte("RIGHT-VALUE")) {
		t.Errorf("conflict must name BOTH values, got left=%q right=%q", c.Left, c.Right)
	}
	if !errors.Is(c, ErrMergeConflict) {
		t.Errorf("conflict report should wrap ErrMergeConflict")
	}

	// No merged value for the conflicted construct: not selected, not
	// concatenated, not interleaved.
	if v, ok := res.Merged[conflicted]; ok {
		t.Errorf("conflicted construct must have NO merged value, got %q", v)
		// Guard against silent concatenation/interleaving too.
		if bytes.Contains(v, []byte("LEFT")) && bytes.Contains(v, []byte("RIGHT")) {
			t.Error("merge must not concatenate/interleave conflicting values")
		}
	}

	// Non-conflicting constructs still merge cleanly.
	if !bytes.Equal(res.Merged[leftOnly], []byte("left-edit")) {
		t.Errorf("one-sided edit should merge, got %q", res.Merged[leftOnly])
	}
	if !bytes.Equal(res.Merged[agreed], []byte("same-new")) {
		t.Errorf("agreed change should merge, got %q", res.Merged[agreed])
	}
}
