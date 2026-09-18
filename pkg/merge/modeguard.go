// History-mode mismatch merge guard (T-0237; CON-025). Two documents/branches
// declaring different immutable history modes cannot both be honoured in one
// merge, so the merge is refused with a named error rather than silently
// choosing one declaration over the other.
package merge

import (
	"errors"
	"fmt"

	"Protodoc/pkg/history"
)

// ErrHistoryModeMismatch is returned when two documents/branches declaring
// different history modes are merged (CON-025).
var ErrHistoryModeMismatch = errors.New("merge: the two inputs declare different history modes")

// HistoryModeMismatchError names both declared modes (CON-025: refuse naming
// both declared modes).
type HistoryModeMismatchError struct {
	ModeA history.Mode
	ModeB history.Mode
}

func (e *HistoryModeMismatchError) Error() string {
	return fmt.Sprintf("%v: %s vs %s", ErrHistoryModeMismatch,
		history.ModeName(e.ModeA), history.ModeName(e.ModeB))
}

func (e *HistoryModeMismatchError) Unwrap() error { return ErrHistoryModeMismatch }

// CheckHistoryModeMatch refuses a merge of two documents/branches whose
// immutable history-mode declarations differ (CON-025). The mode is fixed at
// creation, so a complete-history branch and a no-history branch cannot both
// be honoured; any silent choice violates one document's declaration. On
// mismatch it returns a *HistoryModeMismatchError naming BOTH declared modes;
// on a match it returns nil.
func CheckHistoryModeMatch(a, b history.Declaration) error {
	if a.Mode() != b.Mode() {
		return &HistoryModeMismatchError{ModeA: a.Mode(), ModeB: b.Mode()}
	}
	return nil
}
