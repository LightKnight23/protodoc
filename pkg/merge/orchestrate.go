// Merge orchestrator (T-0241). Composes the M13 pieces into one end-to-end
// three-way merge: precondition guards (duplicate identifier, history-mode
// match, retention point, erased-replay) run first and can refuse the whole
// merge; only if all guards pass does the construct-level three-way merge run,
// reporting cleanly merged constructs and any unresolved conflicts. The
// orchestrator never silently resolves a conflict.
package merge

import (
	"Protodoc/pkg/history"
	"Protodoc/pkg/pdlfmt"
)

// MergeRequest is the full input to an end-to-end three-way merge.
type MergeRequest struct {
	// Common ancestor and the two branches, keyed by construct identity.
	Base  map[pdlfmt.UnitID][]byte
	Left  map[pdlfmt.UnitID][]byte
	Right map[pdlfmt.UnitID][]byte

	// History-mode declarations of the two branches (CON-025).
	LeftDecl  history.Declaration
	RightDecl history.Declaration

	// Left/Right units with identities and digests for the duplicate-identifier
	// guard (FR-024), and their source names.
	LeftUnits  []MergeUnit
	RightUnits []MergeUnit

	// Trimmed retained-from-point declaration (if any) and the incoming
	// branch's change ordinals, for the retention-point guard (CON-024). If
	// TrimmedDecl's mode is not retained-from-point the guard is inert.
	TrimmedDecl      history.Declaration
	IncomingOrdinals []uint16

	// Erased units (trimmed erasure records) and the incoming changes, for the
	// erased-replay guard (FR-096).
	Erased          []history.ErasureRecord
	IncomingChanges []IncomingChange
}

// MergeResult is the orchestrator's outcome: either a refusal (Refused set,
// with the reason) before any merge ran, or a completed three-way merge (Result
// set) which may itself carry unresolved conflicts.
type MergeResult struct {
	Refused error          // non-nil if a precondition refused the whole merge
	Result  ThreeWayResult // valid only when Refused == nil
}

// Orchestrate runs the end-to-end merge. It applies the precondition guards in
// order; the first failing guard refuses the entire merge (no partial output).
// Only when every guard passes does it run the construct-level three-way merge.
func Orchestrate(req MergeRequest) MergeResult {
	// Guard 1: history-mode match (CON-025).
	if err := CheckHistoryModeMatch(req.LeftDecl, req.RightDecl); err != nil {
		return MergeResult{Refused: err}
	}
	// Guard 2: no merge across a retention point (CON-024).
	if err := CheckRetentionPoint(req.TrimmedDecl, req.IncomingOrdinals); err != nil {
		return MergeResult{Refused: err}
	}
	// Guard 3: no duplicate identifier across the two inputs (FR-024).
	if err := CheckDuplicateIdentifier(req.LeftUnits, req.RightUnits); err != nil {
		return MergeResult{Refused: err}
	}
	// Guard 4: no replay of an erased unit (FR-096).
	idx := NewErasureIndex(req.Erased)
	for _, c := range req.IncomingChanges {
		if err := idx.CheckReplayOfErasedUnit(c); err != nil {
			return MergeResult{Refused: err}
		}
	}

	// All guards passed: run the construct-level three-way merge (TR-002/TR-003).
	return MergeResult{Result: ThreeWayMerge(req.Base, req.Left, req.Right)}
}
