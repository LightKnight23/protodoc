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

// ErrMissingPredecessor is returned when a transmitted change's causal
// predecessor is not held by the reader (FR-095).
var ErrMissingPredecessor = errors.New("merge: transmitted change's causal predecessor is not held")

// MissingPredecessorError names the missing predecessor's identifier (FR-095:
// report the missing predecessor's identifier).
type MissingPredecessorError struct {
	Predecessor [32]byte // the missing causal predecessor state-id
}

func (e *MissingPredecessorError) Error() string {
	return fmt.Sprintf("%v: missing predecessor %x", ErrMissingPredecessor, e.Predecessor)
}

func (e *MissingPredecessorError) Unwrap() error { return ErrMissingPredecessor }

// PredecessorDisposition is what a reader does with a change whose predecessor
// it does not hold: buffer it (hold for later) or refuse it. Both are
// conforming (FR-095: "buffer OR refuse"); neither applies the change.
type PredecessorDisposition int

const (
	// Buffered: the change is held pending its predecessor's arrival.
	Buffered PredecessorDisposition = iota
	// Refused: the change is rejected outright.
	Refused
	// Applicable: the predecessor is held; the change may be applied.
	Applicable
)

// zeroState is the all-zero state-id (a root change with no predecessor).
var zeroState [32]byte

// CheckCausalPredecessor decides the disposition of a transmitted change with
// the given causal predecessor state-id, against the set of states the reader
// holds. A change whose predecessor is held (or whose predecessor is the zero
// root) is Applicable. Otherwise the reader must NOT apply it: it returns the
// caller's chosen non-applying disposition (buffer or refuse) together with a
// *MissingPredecessorError naming the missing predecessor. `bufferPolicy`
// selects buffering (true) or refusal (false); either satisfies FR-095.
func CheckCausalPredecessor(predecessor [32]byte, held map[[32]byte]bool, bufferPolicy bool) (PredecessorDisposition, error) {
	if predecessor == zeroState || held[predecessor] {
		return Applicable, nil
	}
	err := &MissingPredecessorError{Predecessor: predecessor}
	if bufferPolicy {
		return Buffered, err
	}
	return Refused, err
}

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
