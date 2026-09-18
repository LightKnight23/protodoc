// Erased-replay refusal (T-0235; FR-096). A merge must refuse a change that
// would re-derive content whose unit has been erased, report the erased unit's
// identity, and persist nothing -- otherwise erasure could be undone by
// replaying an old change.
package merge

import (
	"errors"
	"fmt"

	"Protodoc/pkg/history"
	"Protodoc/pkg/pdlfmt"
)

// ErrReplayOfErasedUnit is returned when an incoming change would re-derive an
// erased unit's content (FR-096).
var ErrReplayOfErasedUnit = errors.New("merge: change would re-derive erased content")

// ReplayOfErasedUnitError names the erased unit whose content the refused
// change would re-derive (FR-096: report the erased unit's identity).
type ReplayOfErasedUnitError struct {
	Erased pdlfmt.UnitID
}

func (e *ReplayOfErasedUnitError) Error() string {
	return fmt.Sprintf("%v: erased unit %x", ErrReplayOfErasedUnit, e.Erased)
}

func (e *ReplayOfErasedUnitError) Unwrap() error { return ErrReplayOfErasedUnit }

// ErasureIndex holds the persistent (trimmed) erasure records the reader knows
// about, indexed for replay detection. The salt is destroyed in the trimmed
// records, so detection matches on identity and the pre-severance digest that
// a replayed change would necessarily reproduce.
type ErasureIndex struct {
	byIdentity map[pdlfmt.UnitID]pdlfmt.Digest256
}

// NewErasureIndex builds a replay-detection index from trimmed erasure records.
func NewErasureIndex(records []history.ErasureRecord) *ErasureIndex {
	idx := &ErasureIndex{byIdentity: make(map[pdlfmt.UnitID]pdlfmt.Digest256, len(records))}
	for _, r := range records {
		idx.byIdentity[r.Identity] = r.Digest
	}
	return idx
}

// IncomingChange is a change offered to the merge: the unit identity it targets
// and the content digest it would re-derive.
type IncomingChange struct {
	Identity pdlfmt.UnitID
	Digest   pdlfmt.Digest256
}

// CheckReplayOfErasedUnit refuses a change that would re-derive an erased
// unit's content: the change targets an erased identity AND reproduces its
// pre-severance digest. It returns a *ReplayOfErasedUnitError naming the erased
// unit; the caller persists nothing. A change to an erased identity that
// carries a DIFFERENT digest is a new, distinct authoring act (not a replay)
// and is permitted -- erasure removes past content, it does not permanently
// forbid the identity.
func (idx *ErasureIndex) CheckReplayOfErasedUnit(c IncomingChange) error {
	if d, ok := idx.byIdentity[c.Identity]; ok && d == c.Digest {
		return &ReplayOfErasedUnitError{Erased: c.Identity}
	}
	return nil
}
