package merge

import (
	"errors"
	"testing"

	"Protodoc/pkg/history"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_096_RefuseReplayOfErasedUnit is T-0235's named integration test
// (FR-096). It reuses history.ErasureRecord: a unit is erased (its salted
// commitment kept, salt trimmed), then a change that would re-derive that
// unit's exact prior content is offered to the merge. The merge refuses it,
// reports the erased unit's identity, and persists nothing. A change to the
// same identity with different content is a new authoring act and is
// permitted.
func TestFR_096_RefuseReplayOfErasedUnit(t *testing.T) {
	erasedID := mTarget(0x07)
	var digest pdlfmt.Digest256
	for i := range digest {
		digest[i] = byte(i + 1)
	}
	var salt [history.SaltSize]byte
	for i := range salt {
		salt[i] = 0xEE
	}

	// Erase the unit, then persist its trimmed form (salt destroyed).
	rec := history.NewErasureRecord(erasedID, digest, salt).Trimmed()
	if !rec.SaltDestroyed() {
		t.Fatal("trimmed erasure record must have its salt destroyed")
	}
	idx := NewErasureIndex([]history.ErasureRecord{rec})

	// A replay: a change re-deriving the erased unit's exact content -> refused.
	persisted := map[pdlfmt.UnitID]bool{}
	replay := IncomingChange{Identity: erasedID, Digest: digest}
	err := idx.CheckReplayOfErasedUnit(replay)
	if !errors.Is(err, ErrReplayOfErasedUnit) {
		t.Fatalf("replay: err = %v, want ErrReplayOfErasedUnit", err)
	}
	var re *ReplayOfErasedUnitError
	if !errors.As(err, &re) || re.Erased != erasedID {
		t.Fatalf("error must name erased unit %x, got %v", erasedID, err)
	}
	// Persist nothing on refusal.
	if err == nil {
		persisted[replay.Identity] = true
	}
	if len(persisted) != 0 {
		t.Errorf("a refused replay must persist nothing, persisted %d units", len(persisted))
	}

	// A new authoring act on the same identity (different digest) -> permitted.
	var other pdlfmt.Digest256
	other[0] = 0xAB
	if err := idx.CheckReplayOfErasedUnit(IncomingChange{Identity: erasedID, Digest: other}); err != nil {
		t.Errorf("new content on an erased identity should be permitted, got %v", err)
	}

	// A change to an unrelated identity -> permitted.
	if err := idx.CheckReplayOfErasedUnit(IncomingChange{Identity: mTarget(0x08), Digest: digest}); err != nil {
		t.Errorf("unrelated identity should be permitted, got %v", err)
	}
}
