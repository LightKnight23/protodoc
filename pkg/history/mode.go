// Package history implements Protodoc's edit-history layer: the closed
// history-mode enum, the RLE HISTORY_OP_BATCH operation log, predecessor-chain
// / nearest-common-ancestor walking, complete-history state reconstruction,
// retention points, and ErasureRecords for states no longer reconstructable.
// It builds on pkg/content and pkg/integrity (for the severance commitment).
package history

import (
	"errors"

	"Protodoc/pkg/container"
)

// Mode is the history-retention mode a document declares at creation
// (CON-022). It is the container package's closed 3-value HistoryMode, reused
// here so there is exactly one such type in the codebase.
type Mode = container.HistoryMode

const (
	// ModeComplete: every previously published state is reconstructable
	// octet-for-octet from the current file (the complete-history reconstruction
	// guarantee, exercised by T-0212).
	ModeComplete = container.HistoryComplete
	// ModeRetainedFromPoint: every state at or after a declared retention
	// point is reconstructable (the retained-from-point guarantee, T-0214).
	ModeRetainedFromPoint = container.HistoryRetainedFromPoint
	// ModeNone: only the current state is retained (CON-022).
	ModeNone = container.HistoryNone
)

// AllModes is the CLOSED set of the three (and only three) history modes
// (CON-022). A value outside it is rejected, never defaulted.
var AllModes = []Mode{ModeComplete, ModeRetainedFromPoint, ModeNone}

// ErrInvalidMode is returned for a history-mode value outside the closed set.
var ErrInvalidMode = errors.New("history: history-mode outside the closed set {complete, retained-from-point, no-history}")

// ValidMode reports whether m is one of the three defined history modes.
func ValidMode(m Mode) bool {
	return m == ModeComplete || m == ModeRetainedFromPoint || m == ModeNone
}

// ModeName returns the canonical spelling of a history mode.
func ModeName(m Mode) string {
	switch m {
	case ModeComplete:
		return "complete-history"
	case ModeRetainedFromPoint:
		return "history-retained-from-point"
	case ModeNone:
		return "no-history"
	default:
		return "invalid"
	}
}

// Declaration is a document's IMMUTABLE history-mode declaration, fixed at
// creation (CON-022). Once constructed it is never changed: there is no
// mutator, and any attempt to declare a second, different mode for the same
// document is refused by CheckImmutable.
type Declaration struct {
	mode Mode
	// retentionPoint is meaningful only when mode is ModeRetainedFromPoint
	// (the segment ordinal at or after which history is retained); 0 otherwise.
	retentionPoint uint16
}

// ErrModeImmutable is returned when a second, different history-mode
// declaration is attempted on a document that already declared one.
var ErrModeImmutable = errors.New("history: history-mode is immutable after creation; a different mode cannot be declared")

// NewDeclaration builds an immutable history-mode declaration. It errors for an
// invalid mode. A retention point is accepted only for ModeRetainedFromPoint
// (0 is forced otherwise); the gating detail is refined in T-0213.
func NewDeclaration(mode Mode) (Declaration, error) {
	if !ValidMode(mode) {
		return Declaration{}, ErrInvalidMode
	}
	return Declaration{mode: mode}, nil
}

// Mode returns the declared history mode.
func (d Declaration) Mode() Mode { return d.mode }

// CheckImmutable enforces CON-022's immutability: redeclaring the SAME mode is
// a no-op (idempotent), but declaring a DIFFERENT mode is refused, naming the
// already-declared mode. This models the "immutable after creation" rule as a
// guard a writer applies before any attempted mode change.
func (d Declaration) CheckImmutable(newMode Mode) error {
	if newMode == d.mode {
		return nil
	}
	return ErrModeImmutable
}
