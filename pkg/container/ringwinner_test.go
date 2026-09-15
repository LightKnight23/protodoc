package container

import (
	"errors"
	"reflect"
	"testing"
)

const selectWinnerTestFileLength = CommitRingSize + 1<<20

// TestFR_117_PDRING001_EqualSequenceTie is T-0008's named conformance
// vector: two ring slots share the highest valid sequence value. Per
// contracts/container.abnf S3.1 this is a structural reject naming both
// slot indices, never resolved by slot position, and must resolve
// identically across repeated runs and independent invocations.
// Implements: FR-117, CON-010.
func TestFR_117_PDRING001_EqualSequenceTie(t *testing.T) {
	var ring [CommitRingSlots]CommitRingRecord
	for i := range ring {
		ring[i] = fixtureRingRecord(uint64(i+1), StateID{})
	}
	const tiedSequence = 100
	ring[3].Sequence = tiedSequence
	ring[5].Sequence = tiedSequence

	enc := EncodeCommitRing(&ring, nil)

	run := func() *RingWinnerTieError {
		_, index, err := SelectWinner(enc, selectWinnerTestFileLength)
		if err == nil {
			t.Fatalf("SelectWinner: got winner at index %d, want PD-RING-001 tie rejection", index)
		}
		var tieErr *RingWinnerTieError
		if !errors.As(err, &tieErr) {
			t.Fatalf("SelectWinner: got %v, want *RingWinnerTieError", err)
		}
		return tieErr
	}

	tie1 := run()
	tie2 := run() // independent invocation: must reproduce the identical verdict

	wantIndices := []int{3, 5}
	if tie1.Sequence != tiedSequence {
		t.Errorf("run 1: tie sequence = %d, want %d", tie1.Sequence, tiedSequence)
	}
	if !reflect.DeepEqual(tie1.SlotIndices, wantIndices) {
		t.Errorf("run 1: slot indices = %v, want %v", tie1.SlotIndices, wantIndices)
	}
	if !reflect.DeepEqual(tie1, tie2) {
		t.Errorf("repeated invocation diverged: run 1 = %+v, run 2 = %+v", tie1, tie2)
	}
}

func TestSelectWinner_PicksHighestValidSequence(t *testing.T) {
	var ring [CommitRingSlots]CommitRingRecord
	for i := range ring {
		ring[i] = fixtureRingRecord(uint64(i+1), StateID{})
	}
	enc := EncodeCommitRing(&ring, nil)

	winner, index, err := SelectWinner(enc, selectWinnerTestFileLength)
	if err != nil {
		t.Fatalf("SelectWinner: %v", err)
	}
	if index != CommitRingSlots-1 {
		t.Errorf("winner index = %d, want %d", index, CommitRingSlots-1)
	}
	if winner.Sequence != CommitRingSlots {
		t.Errorf("winner sequence = %d, want %d", winner.Sequence, CommitRingSlots)
	}
}

func TestSelectWinner_SkipsSelfDigestInvalidSlot(t *testing.T) {
	var ring [CommitRingSlots]CommitRingRecord
	for i := range ring {
		ring[i] = fixtureRingRecord(uint64(i+1), StateID{})
	}
	enc := EncodeCommitRing(&ring, nil)

	// Corrupt the highest-sequence slot's stored octets so its
	// record-digest no longer verifies; it must lose to the next-highest
	// otherwise-eligible slot, not be treated as absent-but-still-winning.
	enc[(CommitRingSlots-1)*RingSlotSize+10] ^= 0xFF

	winner, index, err := SelectWinner(enc, selectWinnerTestFileLength)
	if err != nil {
		t.Fatalf("SelectWinner: %v", err)
	}
	if index != CommitRingSlots-2 {
		t.Errorf("winner index = %d, want %d", index, CommitRingSlots-2)
	}
	if winner.Sequence != CommitRingSlots-1 {
		t.Errorf("winner sequence = %d, want %d", winner.Sequence, CommitRingSlots-1)
	}
}

func TestSelectWinner_ExcludesLedgerLengthExceedingFile(t *testing.T) {
	var ring [CommitRingSlots]CommitRingRecord
	for i := range ring {
		ring[i] = fixtureRingRecord(uint64(i+1), StateID{})
	}
	enc := EncodeCommitRing(&ring, nil)

	// A file length shorter than the highest-sequence slot's own claimed
	// ledger-length must exclude it (truncation/rollback protection),
	// leaving the next-highest slot (whose ledger-length still fits) as
	// winner.
	fileLength := ring[CommitRingSlots-1].LedgerLength - 1

	winner, index, err := SelectWinner(enc, fileLength)
	if err != nil {
		t.Fatalf("SelectWinner: %v", err)
	}
	if index != CommitRingSlots-2 {
		t.Errorf("winner index = %d, want %d", index, CommitRingSlots-2)
	}
	if winner.Sequence != CommitRingSlots-1 {
		t.Errorf("winner sequence = %d, want %d", winner.Sequence, CommitRingSlots-1)
	}
}

func TestSelectWinner_NoEligibleSlots(t *testing.T) {
	var ring [CommitRingSlots]CommitRingRecord // all-zero: never written
	enc := EncodeCommitRing(&ring, nil)
	for i := range ring {
		// Zero out each slot's freshly computed digest so every slot is
		// indistinguishable from a genuinely never-written (all-zero) one.
		copy(enc[i*RingSlotSize+roRecordDigest:(i+1)*RingSlotSize], make([]byte, 32))
	}

	_, index, err := SelectWinner(enc, selectWinnerTestFileLength)
	if !errors.Is(err, ErrNoValidRingSlot) {
		t.Fatalf("SelectWinner: got (%v, %v), want ErrNoValidRingSlot", index, err)
	}
}

func TestSelectWinner_RejectsTruncatedRing(t *testing.T) {
	_, _, err := SelectWinner(make([]byte, CommitRingSize-1), selectWinnerTestFileLength)
	if !errors.Is(err, ErrCommitRingTruncated) {
		t.Fatalf("got %v, want ErrCommitRingTruncated", err)
	}
}
