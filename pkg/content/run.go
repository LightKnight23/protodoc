// Run and base_ordinal positioning (T-0068, CON-001). A Run is the unit of
// persisted character identity (data-model.md S2.8): NFC text carrying a
// minted run_id, whose characters are addressed by (run_id, base_ordinal +
// offset). base_ordinal counts Unicode SCALAR VALUES within one addressable
// text segment only (CON-001): it is a run-internal, reader-recomputable
// offset, never a document-wide persisted counted position. Per the
// A-FIELD-ROLE vocabulary both run_id and base_ordinal are
// IDENTITY-COMPONENT: comparable and look-up-able, never usable for
// positional arithmetic measured from a sequence start.
//
// Consequently nothing in this package computes or returns a cross-segment
// absolute character count: scalar-value counting is always scoped to a
// single run's own text.
package content

import (
	"unicode/utf8"

	"Protodoc/pkg/pdlfmt"
)

// Run is one identity-bearing text segment. BaseOrdinal is the scalar-value
// offset, within this run's own segment, of the run's first character; a
// character at index i (0-based, in scalar values) within Text has identity
// (RunID, BaseOrdinal+i). BaseOrdinal is authoritative when persisted but a
// reader may recompute it; it is never a document-relative counted position.
type Run struct {
	RunID       pdlfmt.UnitID
	BaseOrdinal uint32
	Text        string // UTF-8, NFC (normalisation enforced by T-0073/T-0074)
	// LangRef is the mandatory language-tag reference for this run's text
	// (FR-031). It resolves to exactly one language tag; the zero value
	// (LangUnset) means "not declared" and is invalid for a persisted run
	// (NewRun and the decode path reject it).
	LangRef LangRef
}

// ScalarLen returns the number of Unicode scalar values in the run's text.
// This is the run's length in the CON-001 counting unit (scalar values),
// scoped entirely to this run: it is the only length notion the content
// package exposes for text, and it counts within one segment, never across
// segments. Invalid UTF-8 bytes each count as one scalar value (utf8's
// RuneCountInString replacement-character behaviour), so a malformed run
// still has a well-defined, bounded length rather than a panic.
func (r Run) ScalarLen() int {
	return utf8.RuneCountInString(r.Text)
}

// CharIdentityAt returns the character identity of the scalar value at
// run-internal index i: the pair (RunID, BaseOrdinal+i). i is a run-scoped
// scalar-value index in [0, ScalarLen()); it is NOT a document-wide
// position. It returns false if i is out of range for this run. The
// returned ordinal is an IDENTITY-COMPONENT, usable for equality and
// lookup, never for measuring distance from a document start.
func (r Run) CharIdentityAt(i int) (runID pdlfmt.UnitID, ordinal uint32, ok bool) {
	if i < 0 || i >= r.ScalarLen() {
		return pdlfmt.UnitID{}, 0, false
	}
	return r.RunID, r.BaseOrdinal + uint32(i), true
}

// EndOrdinal returns the scalar-value ordinal one past this run's last
// character within its own segment: BaseOrdinal + ScalarLen(). It is used by
// the merge predicate (T-0070) to test base_ordinal-contiguity of two
// adjacent runs; it is a run-internal ordinal, not a document position.
func (r Run) EndOrdinal() uint32 {
	return r.BaseOrdinal + uint32(r.ScalarLen())
}

// scalarByteOffset returns the byte offset within text at which the
// scalar-value index i begins, and whether i is a valid boundary in
// [0, ScalarLen()]. It is the run-internal mapping from a scalar-value index
// to a byte position used by SplitRun (T-0069); it never escapes this
// package as a returned value, so it introduces no persisted or exported
// counted position.
func scalarByteOffset(text string, i int) (int, bool) {
	if i < 0 {
		return 0, false
	}
	count := 0
	for b := range text { // ranging a string yields the byte index of each rune start
		if count == i {
			return b, true
		}
		count++
	}
	if count == i {
		return len(text), true // i == ScalarLen(): the end boundary
	}
	return 0, false
}

// SplitRun splits r at scalar-value index at (0 <= at <= r.ScalarLen()),
// returning the left piece [0,at) and the right piece [at,end). Both pieces
// share r's run_id -- splitting a run never mints a new identity (FR-020,
// DP-003/DP-004): only base_ordinal shifts. The left piece keeps r's
// BaseOrdinal; the right piece's BaseOrdinal is r.BaseOrdinal+at, so a
// character's identity (run_id, base_ordinal+offset) is invariant across the
// split. at is a run-internal scalar-value index, never a document position.
//
// A split at 0 or at ScalarLen() is legal and yields one empty piece and one
// equal-to-r piece (identity still preserved); at outside [0,ScalarLen()] is
// rejected. Splitting is exactly invertible by MergeRuns (T-0070).
func SplitRun(r Run, at int) (left, right Run, ok bool) {
	byteAt, valid := scalarByteOffset(r.Text, at)
	if !valid {
		return Run{}, Run{}, false
	}
	left = Run{
		RunID:       r.RunID,
		BaseOrdinal: r.BaseOrdinal,
		Text:        r.Text[:byteAt],
	}
	right = Run{
		RunID:       r.RunID,
		BaseOrdinal: r.BaseOrdinal + uint32(at),
		Text:        r.Text[byteAt:],
	}
	return left, right, true
}

// CanMergeRuns reports whether adjacent runs a and b can merge back into
// one: they must share run_id and be base_ordinal-contiguous, i.e.
// b.BaseOrdinal == a.BaseOrdinal + a.ScalarLen() (a's end ordinal). This is
// a pure syntactic predicate over the two runs' identity lineage with no
// side channel: it inspects only run_id and base_ordinal arithmetic, never
// content, actor, clock or session. It is the precondition MergeRuns
// enforces and the exact condition SplitRun's two outputs satisfy.
func CanMergeRuns(a, b Run) bool {
	return a.RunID.Equal(b.RunID) && b.BaseOrdinal == a.EndOrdinal()
}

// MergeRuns merges adjacent runs a and b into one run iff CanMergeRuns(a,b),
// returning the merged run and true, or the zero run and false otherwise.
// The merged run keeps a's run_id and BaseOrdinal and concatenates the two
// texts. MergeRuns is the exact left-inverse of SplitRun: for any run r and
// valid split point at, MergeRuns(SplitRun(r, at)) reconstructs r exactly.
func MergeRuns(a, b Run) (Run, bool) {
	if !CanMergeRuns(a, b) {
		return Run{}, false
	}
	return Run{
		RunID:       a.RunID,
		BaseOrdinal: a.BaseOrdinal,
		Text:        a.Text + b.Text,
	}, true
}

// MoveRun returns a new run sequence with the run at index from relocated to
// index to, preserving the relative order of the others. Moving a run
// changes only its position in the traversal/storage order (external
// positional metadata); it never alters any run's run_id, base_ordinal or
// text (FR-020). The input slice is not mutated. from and to must be valid
// indices; otherwise ok is false.
func MoveRun(runs []Run, from, to int) (out []Run, ok bool) {
	n := len(runs)
	if from < 0 || from >= n || to < 0 || to >= n {
		return nil, false
	}
	out = make([]Run, 0, n)
	out = append(out, runs[:from]...)
	out = append(out, runs[from+1:]...) // remove the moved run
	// Insert it at to (relative to the post-removal slice, to is an index
	// into the original sequence; clamp into the rebuilt slice).
	moved := runs[from]
	tail := append([]Run(nil), out[to:]...)
	out = append(out[:to], moved)
	out = append(out, tail...)
	return out, true
}

// ReorderRuns returns a new run sequence permuted by perm: out[i] =
// runs[perm[i]]. perm must be a permutation of [0,len(runs)); otherwise ok
// is false. Reordering changes only traversal order, never any run's
// run_id, base_ordinal or text (FR-020). The input slice is not mutated.
func ReorderRuns(runs []Run, perm []int) (out []Run, ok bool) {
	n := len(runs)
	if len(perm) != n {
		return nil, false
	}
	seen := make([]bool, n)
	for _, p := range perm {
		if p < 0 || p >= n || seen[p] {
			return nil, false // not a valid permutation
		}
		seen[p] = true
	}
	out = make([]Run, n)
	for i, p := range perm {
		out[i] = runs[p]
	}
	return out, true
}
