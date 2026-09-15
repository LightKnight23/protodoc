// CommitRing winner selection: PD-RING-001 (contracts/container.abnf S3.1,
// data-model.md CommitRing invariants 2-3, plan.md Data flow "Open" step 4).
package container

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
)

// ringSlotDigestValid reports whether slot's stored record-digest (its
// final 32 octets) matches SHA-256 recomputed over the slot's own
// preceding 480 octets (contracts/container.abnf S3 record-digest
// NORMATIVE comment: "Verified before any other ring-slot field is
// trusted"). T-0009 promotes this check to a public, independently
// tested capability (VerifyRingSlotDigest); SelectWinner uses it here to
// determine PD-RING-001 eligibility.
func ringSlotDigestValid(slot []byte) bool {
	if len(slot) < RingSlotSize {
		return false
	}
	slot = slot[:RingSlotSize]
	got := sha256.Sum256(slot[:roRecordDigest])
	return bytes.Equal(got[:], slot[roRecordDigest:RingSlotSize])
}

// ErrNoValidRingSlot is returned when no ring slot has both a verifying
// record-digest and a ledger-length not exceeding the actual file length.
var ErrNoValidRingSlot = errors.New("container: no ring slot has a valid, length-consistent record-digest")

// RingWinnerTieError is PD-RING-001's structural rejection: two or more
// ring slots share the highest eligible sequence value. Per contracts/
// container.abnf S3.1, this is an INVALID verdict, never resolved by slot
// position or wall-clock.
type RingWinnerTieError struct {
	Sequence    uint64
	SlotIndices []int
}

func (e *RingWinnerTieError) Error() string {
	return fmt.Sprintf("container: ring winner tie at sequence %d across slots %v (PD-RING-001)", e.Sequence, e.SlotIndices)
}

// SelectWinner implements PD-RING-001. src must hold at least
// CommitRingSize raw octets (the 7 ring slots in slot order). Over the
// slots whose record-digest verifies AND whose ledger-length does not
// exceed fileLength, it selects the slot with the strictly highest
// sequence value. Two or more eligible slots sharing that highest
// sequence is a structural reject (*RingWinnerTieError) naming every
// tied slot index in ascending order, deterministic across repeated
// calls: this function reads no ambient state (no wall-clock, no map
// iteration) and slot order is fixed at 0..6 (CP-004).
func SelectWinner(src []byte, fileLength uint64) (winner CommitRingRecord, index int, err error) {
	if len(src) < CommitRingSize {
		return CommitRingRecord{}, -1, ErrCommitRingTruncated
	}

	type candidate struct {
		index  int
		record CommitRingRecord
	}
	var eligible []candidate
	for i := 0; i < CommitRingSlots; i++ {
		slot := src[i*RingSlotSize : (i+1)*RingSlotSize]
		if !ringSlotDigestValid(slot) {
			continue
		}
		rec, decErr := DecodeCommitRingRecord(slot)
		if decErr != nil {
			return CommitRingRecord{}, -1, fmt.Errorf("container: ring slot %d: %w", i, decErr)
		}
		if rec.LedgerLength > fileLength {
			continue
		}
		eligible = append(eligible, candidate{index: i, record: *rec})
	}

	if len(eligible) == 0 {
		return CommitRingRecord{}, -1, ErrNoValidRingSlot
	}

	var highest uint64
	for _, c := range eligible {
		if c.record.Sequence > highest {
			highest = c.record.Sequence
		}
	}

	var winners []candidate
	for _, c := range eligible {
		if c.record.Sequence == highest {
			winners = append(winners, c)
		}
	}

	if len(winners) > 1 {
		indices := make([]int, len(winners))
		for i, w := range winners {
			indices[i] = w.index
		}
		return CommitRingRecord{}, -1, &RingWinnerTieError{Sequence: highest, SlotIndices: indices}
	}

	return winners[0].record, winners[0].index, nil
}
