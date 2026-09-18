package merge

import (
	"bytes"
	"errors"
	"testing"

	"Protodoc/pkg/history"
	"Protodoc/pkg/pdlfmt"
)

// TestM13_MergeOrchestratorEndToEnd is T-0241's named integration test. It
// drives the whole M13 pipeline: a clean merge that passes every guard and
// resolves construct-level changes (including a reported conflict), plus each
// precondition guard refusing the whole merge with no output.
func TestM13_MergeOrchestratorEndToEnd(t *testing.T) {
	complete, _ := history.NewDeclaration(history.ModeComplete)
	none, _ := history.NewDeclaration(history.ModeNone)

	base := map[pdlfmt.UnitID][]byte{
		mTarget(0x01): []byte("intro"),
		mTarget(0x02): []byte("body"),
	}
	left := map[pdlfmt.UnitID][]byte{
		mTarget(0x01): []byte("intro-left"), // one-sided edit
		mTarget(0x02): []byte("body"),
	}
	right := map[pdlfmt.UnitID][]byte{
		mTarget(0x01): []byte("intro-right"), // conflicts with left
		mTarget(0x02): []byte("body"),
	}

	// Happy path: all guards pass; merge runs and reports the conflict.
	res := Orchestrate(MergeRequest{
		Base: base, Left: left, Right: right,
		LeftDecl: complete, RightDecl: complete,
		TrimmedDecl: complete, // not retained-from-point -> guard inert
	})
	if res.Refused != nil {
		t.Fatalf("clean merge refused: %v", res.Refused)
	}
	if !res.Result.HasConflict() || len(res.Result.Conflicts) != 1 {
		t.Fatalf("expected exactly 1 conflict, got %+v", res.Result.Conflicts)
	}
	if _, ok := res.Result.Merged[mTarget(0x01)]; ok {
		t.Error("conflicted construct must not be auto-merged")
	}
	if !bytes.Equal(res.Result.Merged[mTarget(0x02)], []byte("body")) {
		t.Error("unchanged construct should carry through")
	}

	// Guard: history-mode mismatch refuses the whole merge (CON-025).
	res = Orchestrate(MergeRequest{
		Base: base, Left: left, Right: right,
		LeftDecl: complete, RightDecl: none, TrimmedDecl: complete,
	})
	if !errors.Is(res.Refused, ErrHistoryModeMismatch) {
		t.Errorf("mode mismatch: refused = %v, want ErrHistoryModeMismatch", res.Refused)
	}

	// Guard: retention-point crossing refuses (CON-024).
	trimmed, _ := history.NewDeclarationWithRetentionPoint(history.ModeRetainedFromPoint, 20)
	res = Orchestrate(MergeRequest{
		Base: base, Left: left, Right: right,
		LeftDecl: complete, RightDecl: complete,
		TrimmedDecl: trimmed, IncomingOrdinals: []uint16{5},
	})
	if !errors.Is(res.Refused, ErrMergeAcrossRetentionPoint) {
		t.Errorf("retention crossing: refused = %v, want ErrMergeAcrossRetentionPoint", res.Refused)
	}

	// Guard: duplicate identifier refuses (FR-024).
	dup := mTarget(0x09)
	res = Orchestrate(MergeRequest{
		Base: base, Left: left, Right: right,
		LeftDecl: complete, RightDecl: complete, TrimmedDecl: complete,
		LeftUnits:  []MergeUnit{{ID: dup, Digest: [32]byte{0xAA}, Source: "L"}},
		RightUnits: []MergeUnit{{ID: dup, Digest: [32]byte{0xBB}, Source: "R"}},
	})
	if !errors.Is(res.Refused, ErrDuplicateIdentifier) {
		t.Errorf("dup id: refused = %v, want ErrDuplicateIdentifier", res.Refused)
	}

	// Guard: erased-unit replay refuses (FR-096).
	erasedID := mTarget(0x0A)
	var dg pdlfmt.Digest256
	dg[0] = 0x77
	var salt [history.SaltSize]byte
	rec := history.NewErasureRecord(erasedID, dg, salt).Trimmed()
	res = Orchestrate(MergeRequest{
		Base: base, Left: left, Right: right,
		LeftDecl: complete, RightDecl: complete, TrimmedDecl: complete,
		Erased:          []history.ErasureRecord{rec},
		IncomingChanges: []IncomingChange{{Identity: erasedID, Digest: dg}},
	})
	if !errors.Is(res.Refused, ErrReplayOfErasedUnit) {
		t.Errorf("erased replay: refused = %v, want ErrReplayOfErasedUnit", res.Refused)
	}
}
