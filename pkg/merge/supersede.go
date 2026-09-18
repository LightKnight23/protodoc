// Supersession recording (T-0238; FR-116). A state that supersedes another
// state which is NOT among its recorded ancestors must record the superseded
// state's identifier together with a resolution value -- otherwise a silent
// last-writer-wins overwrite would erase an author's work undetectably. This
// makes silent overwrite detectable from the file alone.
package merge

import (
	"errors"
	"fmt"

	"Protodoc/pkg/container"
)

// Resolution is how a supersession was resolved. It is a closed enum: a
// supersession is either a deliberate override of the other state or a
// tie-broken selection; both are recorded, neither is silent.
type Resolution uint8

const (
	// ResolutionOverride: the writer deliberately chose this state over the
	// superseded one.
	ResolutionOverride Resolution = iota
	// ResolutionTiebreak: the state won a deterministic tiebreak (e.g. R2
	// total-order by state-id) over the superseded one.
	ResolutionTiebreak
)

// SupersessionRecord records, on a state that supersedes a non-ancestor state,
// the superseded state's identifier and the resolution value (FR-116). Its
// presence makes a would-be silent overwrite detectable from the file alone.
type SupersessionRecord struct {
	Superseded container.StateID // the state this state supersedes
	Resolution Resolution        // how the supersession was resolved
}

// ErrUnrecordedSupersession is returned when a state supersedes a non-ancestor
// without recording it (FR-116).
var ErrUnrecordedSupersession = errors.New("merge: state supersedes a non-ancestor state without recording it")

// UnrecordedSupersessionError names the superseded state whose supersession the
// producing state failed to record (FR-116).
type UnrecordedSupersessionError struct {
	Superseded container.StateID
}

func (e *UnrecordedSupersessionError) Error() string {
	return fmt.Sprintf("%v: superseded state %x unrecorded", ErrUnrecordedSupersession, e.Superseded)
}

func (e *UnrecordedSupersessionError) Unwrap() error { return ErrUnrecordedSupersession }

// RequireSupersessionRecord enforces FR-116: a state that supersedes a state
// which is NOT among its recorded ancestors must carry a SupersessionRecord
// naming that superseded state with a resolution value. `ancestors` is the set
// of the producing state's recorded ancestor ids; `records` are its
// supersession records indexed by superseded id. If the superseded state is an
// ancestor, no record is required (it is ordinary linear progression). If it
// is a non-ancestor, a matching record MUST be present, or the function
// returns an *UnrecordedSupersessionError naming the superseded state.
func RequireSupersessionRecord(superseded container.StateID, ancestors map[container.StateID]bool, records map[container.StateID]SupersessionRecord) error {
	if ancestors[superseded] {
		return nil // superseding an ancestor is ordinary progression, not a conflict
	}
	if _, ok := records[superseded]; !ok {
		return &UnrecordedSupersessionError{Superseded: superseded}
	}
	return nil
}
