package container

import (
	"fmt"
	"testing"
)

// TestFR_117_InterruptedCommitSingleReadableState is T-0010's named
// conformance suite: it simulates a writer killed mid-write, in turn, at
// each of the 7 commit-ring slots, and asserts the file remains readable
// at exactly one complete, self-digest-valid state -- never a tie, never
// a decode error, and never a blend of the pre-commit and post-commit
// content (contracts/container.abnf S3.1 last paragraph). The torn
// slot's own self-digest fails (VerifyRingSlotDigest, T-0009), so
// SelectWinner (T-0008) resolves among the 6 untouched slots.
// Implements: FR-117, CON-010.
func TestFR_117_InterruptedCommitSingleReadableState(t *testing.T) {
	const fileLength = CommitRingSize + 1<<20
	// Mid-slot: neither a no-op (0, nothing written yet) nor a complete
	// write (RingSlotSize, no tear at all) -- a genuine partial write.
	const tornWriteOffset = 256

	var preCommit [CommitRingSlots]CommitRingRecord
	for i := range preCommit {
		preCommit[i] = fixtureRingRecord(uint64(i+1), StateID{})
	}
	preCommitBytes := EncodeCommitRing(&preCommit, nil)

	// The attempted next commit: sequence 8, one past the pre-commit
	// highest.
	newRecord := fixtureRingRecord(uint64(CommitRingSlots+1), StateID{})
	newRecordBytes := newRecord.Encode(nil)

	for target := 0; target < CommitRingSlots; target++ {
		t.Run(fmt.Sprintf("slot%d", target), func(t *testing.T) {
			torn := append([]byte(nil), preCommitBytes...)
			slotStart := target * RingSlotSize
			// A crash partway through overwriting slot `target` with
			// newRecordBytes in place: the first tornWriteOffset octets
			// are the new record's bytes; the remainder is whatever was
			// already on disk (this slot's own pre-commit content) --
			// never bytes from two different in-flight commits.
			copy(torn[slotStart:slotStart+tornWriteOffset], newRecordBytes[:tornWriteOffset])

			tornSlot := torn[slotStart : slotStart+RingSlotSize]
			if VerifyRingSlotDigest(tornSlot) {
				t.Fatalf("slot %d: torn write is unexpectedly still self-digest-valid; fixture does not exercise a real tear", target)
			}

			winner, index, err := SelectWinner(torn, fileLength)
			if err != nil {
				t.Fatalf("slot %d torn: SelectWinner returned %v, want exactly one readable state", target, err)
			}

			// Every slot except `target` still holds its untouched
			// pre-commit content, so the winner is the highest surviving
			// pre-commit sequence: CommitRingSlots unless the torn slot
			// was itself the pre-commit winner (target == last index), in
			// which case the next-highest survivor wins.
			wantIndex := CommitRingSlots - 1
			wantSequence := uint64(CommitRingSlots)
			if target == CommitRingSlots-1 {
				wantIndex = CommitRingSlots - 2
				wantSequence = uint64(CommitRingSlots - 1)
			}
			if index != wantIndex {
				t.Errorf("slot %d torn: winner index = %d, want %d", target, index, wantIndex)
			}
			if winner.Sequence != wantSequence {
				t.Errorf("slot %d torn: winner sequence = %d, want %d", target, winner.Sequence, wantSequence)
			}

			// Re-running selection over the identical bytes must
			// reproduce the identical outcome: no ambiguous or
			// dual-valid outcome.
			winner2, index2, err2 := SelectWinner(torn, fileLength)
			if err2 != nil || index2 != index || winner2 != winner {
				t.Errorf("slot %d torn: repeated selection diverged: first=(%+v,%d,%v) second=(%+v,%d,%v)",
					target, winner, index, err, winner2, index2, err2)
			}
		})
	}
}
