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
