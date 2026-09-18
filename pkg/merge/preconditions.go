// Merge preconditions (T-0233..T-0237). Before merging two documents, a
// conforming implementation checks preconditions that categorically refuse an
// unsafe merge, naming the offending values, rather than silently producing a
// corrupt result. This file starts with FR-024's duplicate-identifier refusal;
// the missing-predecessor, erased-replay, retention-point, and history-mode
// refusals are added by their own tasks (which own their requirement ids).
package merge

import (
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// MergeUnit is one content unit as a merge input sees it: its identifier and a
// content digest distinguishing distinct units that happen to share an id.
type MergeUnit struct {
	ID     pdlfmt.UnitID
	Digest [32]byte
	Source string // the source document this unit came from
}

// ErrDuplicateIdentifier is returned when two DISTINCT content units in the
// merge input pair carry the same identifier (FR-024).
var ErrDuplicateIdentifier = errors.New("merge: two distinct content units carry the same identifier")

// DuplicateIdentifierError names both offending units and their source
// documents (FR-024: refuse naming both units and their source documents).
type DuplicateIdentifierError struct {
	ID      pdlfmt.UnitID
	SourceA string
	SourceB string
}

func (e *DuplicateIdentifierError) Error() string {
	return fmt.Sprintf("%v: id %x appears as distinct units in %q and %q", ErrDuplicateIdentifier, e.ID, e.SourceA, e.SourceB)
}

func (e *DuplicateIdentifierError) Unwrap() error { return ErrDuplicateIdentifier }

// CheckDuplicateIdentifier refuses a merge when a unit in inputA and a unit in
// inputB share an identifier but are DISTINCT (different content digest) --
// two lineages independently minting the same id (FR-024). A shared id with an
// identical digest is the SAME unit (a legitimate common ancestor), not a
// collision, and is permitted. On a collision it returns a
// *DuplicateIdentifierError naming both units and their sources.
func CheckDuplicateIdentifier(inputA, inputB []MergeUnit) error {
	byID := make(map[pdlfmt.UnitID]MergeUnit, len(inputA))
	for _, u := range inputA {
		byID[u.ID] = u
	}
	for _, u := range inputB {
		if a, ok := byID[u.ID]; ok && a.Digest != u.Digest {
			return &DuplicateIdentifierError{ID: u.ID, SourceA: a.Source, SourceB: u.Source}
		}
	}
	return nil
}
