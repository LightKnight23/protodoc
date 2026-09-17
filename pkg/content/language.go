// Mandatory per-span language reference (T-0081, FR-031; data-model.md
// S2.7/S2.8). Every text span -- every TextBlock and every Run -- carries a
// language-tag reference that resolves to exactly one language tag; the
// field is MANDATORY, not optional. A span with no declared language
// reference is invalid: construction rejects it and the decode path rejects
// a record lacking the field before any resolution logic runs. (The
// resolution check itself -- that the reference resolves to a known tag --
// is a validator rule wired by a later task, T-0082.)
package content

import (
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// LangRef is a language-tag reference: a registry-issued identifier that
// resolves to exactly one language tag (a BCP-47 tag in the document's
// registry excerpt). The zero value LangUnset is reserved to mean "not
// declared" and is never a valid persisted reference.
type LangRef uint16

// LangUnset is the reserved zero value meaning no language reference was
// declared; it is invalid for a persisted text span (FR-031).
const LangUnset LangRef = 0

// IsSet reports whether the reference was declared (nonzero).
func (l LangRef) IsSet() bool { return l != LangUnset }

// TextBlock is an addressable text segment: a mandatory language reference
// plus its ordered runs (data-model.md S2.7). Every TextBlock declares
// exactly one language reference for the block; individual runs may carry
// their own reference too, but neither may be unset (FR-031).
type TextBlock struct {
	BlockID pdlfmt.UnitID
	LangRef LangRef
	Runs    []Run
}

var (
	// ErrMissingLanguageRef is returned when a text span is constructed or
	// decoded without a declared language reference (FR-031).
	ErrMissingLanguageRef = errors.New("content: text span has no declared language-tag reference (FR-031)")
)

// NewRun constructs a Run, requiring a declared (nonzero) language
// reference. It returns ErrMissingLanguageRef if lang is LangUnset, so a run
// can never be constructed without a language reference.
func NewRun(runID pdlfmt.UnitID, baseOrdinal uint32, text string, lang LangRef) (Run, error) {
	if !lang.IsSet() {
		return Run{}, fmt.Errorf("%w: run %x", ErrMissingLanguageRef, runID)
	}
	return Run{RunID: runID, BaseOrdinal: baseOrdinal, Text: text, LangRef: lang}, nil
}

// NewTextBlock constructs a TextBlock, requiring a declared (nonzero) block
// language reference and requiring every run to carry a declared reference
// too. It returns ErrMissingLanguageRef naming the offending span otherwise.
func NewTextBlock(blockID pdlfmt.UnitID, lang LangRef, runs []Run) (TextBlock, error) {
	if !lang.IsSet() {
		return TextBlock{}, fmt.Errorf("%w: block %x", ErrMissingLanguageRef, blockID)
	}
	for i, r := range runs {
		if !r.LangRef.IsSet() {
			return TextBlock{}, fmt.Errorf("%w: block %x run %d (%x)", ErrMissingLanguageRef, blockID, i, r.RunID)
		}
	}
	return TextBlock{BlockID: blockID, LangRef: lang, Runs: runs}, nil
}

// ValidateLanguageRefPresent checks that a decoded text span carries a
// declared language reference, rejecting it before any resolution logic
// (FR-031: the field is mandatory). It is the decode-time trust-boundary
// counterpart of the NewRun/NewTextBlock construction check.
func ValidateLanguageRefPresent(spanID pdlfmt.UnitID, lang LangRef) error {
	if !lang.IsSet() {
		return fmt.Errorf("%w: span %x", ErrMissingLanguageRef, spanID)
	}
	return nil
}
