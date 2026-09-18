// Retention-point merge guard (T-0236; CON-024). A merge that would silently
// undo a trim is refused with a named error rather than producing a result
// that violates the trimming party's retention declaration. (The history-mode
// mismatch guard is added by T-0237.)
package merge

import (
	"errors"
	"fmt"

	"Protodoc/pkg/history"
)

// ErrMergeAcrossRetentionPoint is returned when a merge input holds changes
// predating a trimmed document's retention point (CON-024).
var ErrMergeAcrossRetentionPoint = errors.New("merge: input holds changes predating the trimmed document's retention point")

// RetentionPointError names the retention point the refused merge would have
// crossed (CON-024: refuse naming the retention point).
type RetentionPointError struct {
	RetentionPoint   uint16
	OffendingOrdinal uint16 // the earliest pre-trim change ordinal in the input
}

func (e *RetentionPointError) Error() string {
	return fmt.Sprintf("%v: change at ordinal %d predates retention point %d",
		ErrMergeAcrossRetentionPoint, e.OffendingOrdinal, e.RetentionPoint)
}

func (e *RetentionPointError) Unwrap() error { return ErrMergeAcrossRetentionPoint }

// CheckRetentionPoint refuses a merge when the incoming branch carries any
// change whose segment ordinal predates the trimmed (retained-from-point)
// document's retention point (CON-024). Merging such a branch would silently
// restore content the trimming party removed. `incomingOrdinals` are the
// segment ordinals of the changes offered by the other branch. The guard
// applies only to a retained-from-point declaration; complete-history and
// no-history declarations have no retention point to cross. On refusal it
// names the retention point and the offending ordinal.
func CheckRetentionPoint(trimmed history.Declaration, incomingOrdinals []uint16) error {
	if trimmed.Mode() != history.ModeRetainedFromPoint {
		return nil
	}
	rp := trimmed.RetentionPoint()
	for _, ord := range incomingOrdinals {
		if ord < rp {
			return &RetentionPointError{RetentionPoint: rp, OffendingOrdinal: ord}
		}
	}
	return nil
}
