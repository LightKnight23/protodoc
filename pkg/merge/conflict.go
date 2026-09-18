// Three-way-merge conflict reporting (T-0240; TR-003). When a three-way merge
// encounters a conflict, it names BOTH values and neither selects a side nor
// concatenates nor interleaves content -- the merge halts on that construct
// with a reported conflict the CLI turns into a non-zero exit.
package merge

import (
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// ErrMergeConflict is returned when a three-way merge cannot resolve a
// construct without guessing (TR-003).
var ErrMergeConflict = errors.New("merge: three-way merge conflict; not auto-resolved")

// ConflictReport names a single unresolved construct-level conflict, carrying
// BOTH sides' values (TR-003: name both values, do not select/concatenate/
// interleave). Neither Left nor Right is combined into a merged value.
type ConflictReport struct {
	Construct pdlfmt.UnitID
	Base      []byte // the common-ancestor value (nil if none)
	Left      []byte // the left branch's value
	Right     []byte // the right branch's value
}

func (c ConflictReport) Error() string {
	return fmt.Sprintf("%v: construct %x left=%q right=%q", ErrMergeConflict, c.Construct, c.Left, c.Right)
}

func (c ConflictReport) Unwrap() error { return ErrMergeConflict }

// ThreeWayResult is the outcome of a three-way merge: the cleanly merged
// constructs plus every unresolved conflict. When Conflicts is non-empty the
// CLI exits non-zero and emits no merged value for the conflicted constructs.
type ThreeWayResult struct {
	Merged    map[pdlfmt.UnitID][]byte
	Conflicts []ConflictReport
}

// HasConflict reports whether the merge produced any unresolved conflict.
func (r ThreeWayResult) HasConflict() bool { return len(r.Conflicts) > 0 }

// ThreeWayMerge merges left and right against a common base, keyed by
// construct identity. A construct changed on only one side takes that side's
// value; a construct changed identically on both sides takes the agreed value.
// A construct changed to DIFFERENT values on both sides is a conflict: it is
// reported (naming both values) and NO merged value is produced for it -- the
// merge never selects a side, concatenates, or interleaves (TR-003).
func ThreeWayMerge(base, left, right map[pdlfmt.UnitID][]byte) ThreeWayResult {
	res := ThreeWayResult{Merged: map[pdlfmt.UnitID][]byte{}}

	ids := make(map[pdlfmt.UnitID]bool)
	for id := range base {
		ids[id] = true
	}
	for id := range left {
		ids[id] = true
	}
	for id := range right {
		ids[id] = true
	}

	for id := range ids {
		b, hasB := base[id]
		l, hasL := left[id]
		r, hasR := right[id]

		leftChanged := !hasB || !hasL || !bytesEqual(b, l)
		rightChanged := !hasB || !hasR || !bytesEqual(b, r)

		switch {
		case leftChanged && rightChanged:
			if hasL && hasR && bytesEqual(l, r) {
				// Both sides made the SAME change -> agreed, no conflict.
				res.Merged[id] = l
			} else {
				// Divergent change on both sides -> conflict; name both values,
				// produce no merged value for this construct.
				res.Conflicts = append(res.Conflicts, ConflictReport{
					Construct: id, Base: b, Left: l, Right: r,
				})
			}
		case leftChanged:
			if hasL {
				res.Merged[id] = l
			}
		case rightChanged:
			if hasR {
				res.Merged[id] = r
			}
		default:
			// Unchanged on both sides -> carry base value.
			if hasB {
				res.Merged[id] = b
			}
		}
	}
	return res
}
