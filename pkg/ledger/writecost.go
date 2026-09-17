// Per-edit write-cost accounting and enforcement (T-0036, NFR-008,
// plan.md Section 5 "Ledger / Storage" row: "at most 8K + 262144 octets
// total write for a K-octet logical edit"). A commit's write cost is the
// octets it actually puts to storage: the appended sealed-segment octets
// plus the fixed-prefix patch a real commit performs (one CommitRingRecord
// slot, plus one touched SegmentTableSlot per newly placed segment). This
// file bounds that quantity as a hard assertion: a commit whose accounted
// write cost would exceed the budget is refused, not merely measured.
package ledger

import (
	"errors"
	"fmt"

	"Protodoc/pkg/container"
)

// WriteBudgetConstant is the additive term of NFR-008's per-edit
// write-cost bound 8*K + WriteBudgetConstant. It is 262144, the same fixed
// headroom the combined index+integrity per-edit delta ceiling uses (that
// ceiling is enforced in its own right by a later M02 task, T-0043); here
// it is the fixed slack a K-octet edit is allowed for its prefix patch and
// index/integrity churn on top of the content it writes.
const WriteBudgetConstant = 262144

// MaxBudgetedDocumentSize is the document-size ceiling NFR-008's bound is
// stated for: 1073741824 octets (1 GiB). The bound is required to hold for
// documents up to this size.
const MaxBudgetedDocumentSize = 1073741824

// prefixPatchCostPerCommit is the fixed-prefix octets a single commit
// patches independent of how many segments it appends: exactly one
// CommitRingRecord slot is written per commit (plan.md "the next
// CommitRingRecord slot (round-robin by sequence mod 7, 512 octets)").
const prefixPatchCostPerCommit = container.RingSlotSize

// prefixPatchCostPerSegment is the fixed-prefix octets patched per newly
// placed segment: exactly one SegmentTableSlot (48 octets) records the new
// segment's ordinal, offset, length and digest (plan.md "the touched
// SegmentTableSlot entries, 48 octets each").
const prefixPatchCostPerSegment = container.SegmentTableSlotSize

// ErrWriteBudgetExceeded is returned by a budgeted commit when the
// accounted write cost would exceed NFR-008's 8*K + 262144 bound. The
// error names both the accounted cost and the budget so a caller sees by
// how much the commit blew the bound.
var ErrWriteBudgetExceeded = errors.New("ledger: per-edit write cost exceeds the NFR-008 budget (8*K + 262144)")

// ErrDocumentTooLargeForBudget is returned when a budgeted commit is asked
// to operate on a prior document larger than MaxBudgetedDocumentSize: the
// NFR-008 bound is only stated for documents up to 1 GiB, so a larger base
// is outside the guarantee's domain and refused rather than silently
// treated as if the bound applied.
var ErrDocumentTooLargeForBudget = errors.New("ledger: prior document exceeds the 1 GiB size the NFR-008 write-cost budget is stated for")

// commitWriteCost returns the total octets a commit writes to storage for
// a delta that appends appendedOctets octets across newSegmentCount
// segments: the appended segment octets plus the per-commit and
// per-segment fixed-prefix patch. It performs no I/O; it is pure
// arithmetic over already-validated quantities.
func commitWriteCost(appendedOctets, newSegmentCount uint64) uint64 {
	if newSegmentCount == 0 {
		// A no-op commit writes nothing: no ring slot succession, no
		// segment-table patch (NFR-003). This keeps a no-op save's write
		// cost at exactly zero, consistent with T-0035.
		return 0
	}
	return appendedOctets + prefixPatchCostPerCommit + newSegmentCount*prefixPatchCostPerSegment
}

// -------- FR-057: combined index+integrity per-edit delta budget --------
//
// FR-057 bounds the combined change to index and integrity data at 262,144
// octets per edit. The integrity trees (T_S storage, T_C content) land in
// M08; until then this accounting uses the documented worst-case update
// sizes from plan.md Section 5 row 5 as a conservative stand-in, which sum
// to ~26,161 octets (about 10x headroom under the 262,144 ceiling). A full
// end-to-end re-verification against the real integrity-tree implementation
// is a scheduled M08 follow-up (see the tracking note in
// TestFR_057_CombinedIndexIntegrityDeltaBudgetEnforced).

// CombinedDeltaBudget is FR-057's per-edit ceiling on the combined index +
// integrity change: 262,144 octets. It equals WriteBudgetConstant (both are
// the 262,144-octet prefix-scale budget), named separately here to make the
// FR-057 accounting self-documenting rather than reusing the NFR-008 term.
const CombinedDeltaBudget = 262144

// Worst-case per-edit delta component sizes (plan.md Section 5 row 5). Each
// is an upper bound on how many octets that structure's update touches for
// a single edit; the integrity-tree components are the documented M08
// stand-in figures, deliberately generous.
const (
	// indexLeafUpdateOctets bounds the UnitIndexLeaf patch an edit causes
	// (plan.md "index leaf (16,384)").
	indexLeafUpdateOctets = 16384

	// segmentTablePatchOctets is the one touched SegmentTableSlot
	// (plan.md "segment-table slot (48)").
	segmentTablePatchOctets = prefixPatchCostPerSegment // 48

	// ringSlotPatchOctets is the one CommitRingRecord slot a commit writes
	// (plan.md "ring slot (512)").
	ringSlotPatchOctets = prefixPatchCostPerCommit // 512

	// tSNodesUpdateOctets is the worst-case T_S storage-tree node path an
	// edit rewrites (plan.md "4 T_S nodes (~2,052)"). M08 stand-in.
	tSNodesUpdateOctets = 2052

	// tCPathUpdateOctets is the worst-case T_C content-tree path an edit
	// rewrites (plan.md "T_C path (~5,165)"). M08 stand-in.
	tCPathUpdateOctets = 5165

	// coverageDescriptorDeltaOctets is the worst-case coverage-descriptor
	// change an edit causes (plan.md "coverage-descriptor delta ~2,000").
	// M08 stand-in.
	coverageDescriptorDeltaOctets = 2000
)

// CombinedDelta is the breakdown of an edit's combined index+integrity
// delta, so a caller (and the FR-057 test) can see each component and the
// total the budget is checked against.
type CombinedDelta struct {
	IndexLeafOctets     uint64
	SegmentTableOctets  uint64
	RingSlotOctets      uint64
	TSNodesOctets       uint64 // integrity-tree stand-in (M08)
	TCPathOctets        uint64 // integrity-tree stand-in (M08)
	CoverageDeltaOctets uint64 // integrity-tree stand-in (M08)
}

// IndexDelta returns the index-side octets (index leaf + segment-table
// slot), the change to the derived-index and inventory data.
func (d CombinedDelta) IndexDelta() uint64 {
	return d.IndexLeafOctets + d.SegmentTableOctets
}

// IntegrityDelta returns the integrity-side octets (T_S nodes + T_C path +
// coverage-descriptor delta), the estimated change to integrity data, plus
// the ring slot that carries the roots.
func (d CombinedDelta) IntegrityDelta() uint64 {
	return d.TSNodesOctets + d.TCPathOctets + d.CoverageDeltaOctets + d.RingSlotOctets
}

// Total returns the combined index+integrity delta FR-057 bounds.
func (d CombinedDelta) Total() uint64 {
	return d.IndexDelta() + d.IntegrityDelta()
}

// WorstCaseCombinedDelta returns the conservative per-edit combined
// index+integrity delta for an edit touching newSegmentCount segments,
// using the plan.md Section 5 row 5 worst-case component figures (the M08
// integrity-tree stand-in). Every component is an upper bound, so the
// returned Total is a ceiling on the real delta a single edit can cause.
func WorstCaseCombinedDelta(newSegmentCount uint64) CombinedDelta {
	if newSegmentCount == 0 {
		return CombinedDelta{}
	}
	return CombinedDelta{
		IndexLeafOctets:     indexLeafUpdateOctets,
		SegmentTableOctets:  newSegmentCount * segmentTablePatchOctets,
		RingSlotOctets:      ringSlotPatchOctets,
		TSNodesOctets:       tSNodesUpdateOctets,
		TCPathOctets:        tCPathUpdateOctets,
		CoverageDeltaOctets: coverageDescriptorDeltaOctets,
	}
}

// ErrCombinedDeltaExceeded is returned when a commit's combined
// index+integrity delta would exceed FR-057's 262,144-octet ceiling.
var ErrCombinedDeltaExceeded = errors.New("ledger: combined index+integrity per-edit delta exceeds the FR-057 ceiling (262144)")

// EnforceCombinedDeltaBudget checks a commit's combined index+integrity
// delta against FR-057's 262,144-octet ceiling, returning
// ErrCombinedDeltaExceeded (naming the total and ceiling) if it would be
// exceeded. It is a hard assertion: a commit over the ceiling is refused,
// not merely measured.
func EnforceCombinedDeltaBudget(delta CombinedDelta) error {
	if total := delta.Total(); total > CombinedDeltaBudget {
		return fmt.Errorf("%w: delta is %d octets, ceiling %d", ErrCombinedDeltaExceeded, total, CombinedDeltaBudget)
	}
	return nil
}

// WriteBudget returns NFR-008's per-edit write-cost bound for a K-octet
// logical edit: 8*K + 262144. logicalEditOctets is K, the count of octets
// of logical content the edit changes.
func WriteBudget(logicalEditOctets uint64) uint64 {
	return 8*logicalEditOctets + WriteBudgetConstant
}

// PlaceWithinBudget performs a place() commit and enforces NFR-008 as a
// hard assertion: given logicalEditOctets (K, the size of the logical edit
// being committed), it refuses the commit if the accounted write cost
// would exceed 8*K + 262144, returning ErrWriteBudgetExceeded and no
// image. It also refuses a prior document larger than the 1 GiB size the
// bound is stated for (ErrDocumentTooLargeForBudget). On success the
// returned PlaceResult carries the same append-only, ordinal-assigning
// guarantees as place(), plus a WrittenOctets known to be within budget.
func PlaceWithinBudget(prior []byte, priorSegmentCount uint64, delta EditDelta, logicalEditOctets uint64) (PlaceResult, error) {
	if uint64(len(prior)) > MaxBudgetedDocumentSize {
		return PlaceResult{}, fmt.Errorf("%w: prior is %d octets", ErrDocumentTooLargeForBudget, len(prior))
	}

	res, err := place(prior, priorSegmentCount, delta)
	if err != nil {
		return PlaceResult{}, err
	}

	budget := WriteBudget(logicalEditOctets)
	if res.WrittenOctets > budget {
		return PlaceResult{}, fmt.Errorf("%w: commit writes %d octets, budget for a %d-octet edit is %d", ErrWriteBudgetExceeded, res.WrittenOctets, logicalEditOctets, budget)
	}
	return res, nil
}
