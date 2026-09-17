// Storage-integrity-tree check (T-0366, FR-104/FR-105; data-model.md S7 step
// 6). cli.md's `validate` stdout schema names a required check key,
// storage_integrity_tree: the validator recomputes T_S fresh from the current
// SegmentTable and compares it against the winning CommitRingRecord's
// ledger_root. A divergence is reported as its own distinct verdict, never
// silently accepted; the stored ledger_root is used ONLY for comparison, never
// trusted as the answer (data-model.md S2.11 note 2). On divergence the check
// names the offending slot -- the storage ordinal whose stored slot-digest no
// longer matches the digest recomputed over its segment octets.
package validate

import (
	"Protodoc/pkg/container"
	"Protodoc/pkg/integrity"
)

// StorageIntegrityCheckKey is the cli.md `validate` stdout check key this
// check reports under.
const StorageIntegrityCheckKey = "storage_integrity_tree"

// StorageIntegrityResult is the outcome of the storage_integrity_tree check.
type StorageIntegrityResult struct {
	// Passed is true iff the freshly recomputed T_S root equals the stored
	// ledger_root.
	Passed bool
	// FreshRoot is the T_S root recomputed from the current SegmentTable.
	FreshRoot integrity.Digest
	// StoredRoot is the winning CommitRingRecord.ledger_root it was compared
	// against (for reporting only; never trusted as authoritative).
	StoredRoot integrity.Digest
	// OffendingSlots lists the storage ordinals whose stored slot-digest does
	// not match the digest recomputed over the slot's segment octets, when the
	// check fails. Empty on a pass. Naming the slot is FR-104/105's
	// requirement that a tamper be localised, not merely detected.
	OffendingSlots []int
}

// SegmentDigester recomputes the SHA-256 digest a slot at storage ordinal
// `ordinal` should carry, over that segment's octets read from the file. It
// returns the digest, or ok=false if the segment cannot be read (e.g. the
// slot is unused). The validate pipeline supplies this from the file under
// test so this package needs no direct file access of its own.
type SegmentDigester func(ordinal int, slot container.SegmentTableSlot) (digest [32]byte, ok bool)

// CheckStorageIntegrityTree recomputes T_S from slots and compares it against
// storedLedgerRoot (the winning CommitRingRecord.ledger_root). It returns a
// StorageIntegrityResult; on a mismatch it uses digester to find the offending
// slot(s) whose stored digest no longer matches their recomputed segment
// digest. digester may be nil, in which case a failing check still reports the
// divergence but with no localised slot.
func CheckStorageIntegrityTree(slots []container.SegmentTableSlot, storedLedgerRoot integrity.Digest, digester SegmentDigester) (StorageIntegrityResult, error) {
	verdict, fresh, err := integrity.CompareLedgerRoot(slots, storedLedgerRoot)
	if err != nil {
		return StorageIntegrityResult{}, err
	}
	res := StorageIntegrityResult{
		Passed:     verdict == integrity.LedgerRootMatches,
		FreshRoot:  fresh,
		StoredRoot: storedLedgerRoot,
	}
	if res.Passed || digester == nil {
		return res, nil
	}
	// Localise: a present slot whose stored digest disagrees with its
	// recomputed segment digest is an offending slot.
	for ord, slot := range slots {
		if slot.SegmentType == container.SegmentTypeUnused {
			continue
		}
		recomputed, ok := digester(ord, slot)
		if !ok {
			continue
		}
		if recomputed != slot.Digest {
			res.OffendingSlots = append(res.OffendingSlots, ord)
		}
	}
	return res, nil
}

// StorageIntegrityFinding maps a failing storage_integrity_tree check to a
// pipeline Finding at step 6 (StepTSRecompute), reporting under the
// storage_integrity_tree check key. Returns nil when the check passed.
func StorageIntegrityFinding(res StorageIntegrityResult) *Finding {
	if res.Passed {
		return nil
	}
	msg := StorageIntegrityCheckKey + ": recomputed T_S root diverges from stored ledger_root"
	if len(res.OffendingSlots) > 0 {
		msg += " (offending storage ordinal(s) detected)"
	}
	return &Finding{
		Step:    StepTSRecompute,
		RuleID:  StorageIntegrityCheckKey,
		Message: msg,
	}
}
