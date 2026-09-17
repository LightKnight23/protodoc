// T_S storage-integrity tree (T-0130, TR-009; integrity.abnf S2.1): a fixed
// arity-16, depth-4 Merkle tree over the SegmentTable. A leaf at tree
// position i is the slot-digest of the segment at storage ordinal i (no
// domain tag of its own); an internal node is 0x01 || 16 child digests,
// ALWAYS a full 513-octet preimage regardless of how many children are real,
// so two implementations hash an identical buffer for an identical logical
// shape. Every tree position with no corresponding segment -- at every level,
// not only the root -- is filled with ABSENT_CHILD_DIGEST.
package integrity

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"Protodoc/pkg/container"
)

// T_S fixed-shape constants (integrity.abnf S2.1).
const (
	// TSArity is the fixed number of children at every T_S internal node.
	TSArity = 16
	// TSDepth is the fixed number of internal levels above the leaves.
	TSDepth = 4
	// TSLeafCapacity is 16^4 = 65536, the leaf capacity (headroom over
	// MAX_SEGMENTS = 16384).
	TSLeafCapacity = 65536
	// tsInternalPreimageLen is 1 + 16*32 = 513 octets, the fixed size of
	// every internal-node preimage.
	tsInternalPreimageLen = 1 + TSArity*32
)

// ErrTSTooManySegments is returned when more segment slots are supplied than
// the depth-4 tree can hold (its 65536-leaf capacity).
var ErrTSTooManySegments = errors.New("integrity: more segment slots than the T_S depth-4 capacity (65536)")

// tsInternalNode hashes an internal node's preimage: 0x01 || the 16 child
// digests concatenated, always exactly 513 octets.
func tsInternalNode(children [TSArity]Digest) Digest {
	var buf [tsInternalPreimageLen]byte
	buf[0] = DomainTSInternal
	for i := 0; i < TSArity; i++ {
		copy(buf[1+i*32:1+(i+1)*32], children[i][:])
	}
	var d Digest
	sum := sha256.Sum256(buf[:])
	copy(d[:], sum[:])
	return d
}

// buildTSTree builds the T_S tree over slots (tree position i == storage
// ordinal i) and returns the root. Leaf position i is slots[i].slot-digest
// if present, else ABSENT_CHILD_DIGEST; the tree is always exactly depth 4
// with a full 65536-leaf base, so absent positions at every level are
// ABSENT_CHILD_DIGEST. It errors if len(slots) exceeds the leaf capacity.
func buildTSTree(slots []container.SegmentTableSlot) (Digest, error) {
	if len(slots) > TSLeafCapacity {
		return Digest{}, fmt.Errorf("%w: got %d", ErrTSTooManySegments, len(slots))
	}

	// Level 0: the 65536 leaves. Present positions take the slot-digest;
	// absent positions take ABSENT_CHILD_DIGEST.
	level := make([]Digest, TSLeafCapacity)
	for i := range level {
		if i < len(slots) {
			level[i] = Digest(slots[i].Digest)
		} else {
			level[i] = AbsentChildDigest
		}
	}

	// Collapse 4 times, grouping every 16 consecutive nodes into one parent.
	for d := 0; d < TSDepth; d++ {
		next := make([]Digest, len(level)/TSArity)
		for p := range next {
			var children [TSArity]Digest
			copy(children[:], level[p*TSArity:(p+1)*TSArity])
			next[p] = tsInternalNode(children)
		}
		level = next
	}

	if len(level) != 1 {
		return Digest{}, fmt.Errorf("integrity: T_S collapse produced %d roots, want 1", len(level))
	}
	return level[0], nil
}

// TSRoot returns the T_S root, ALWAYS recomputed from the current live
// SegmentTable octets (integrity.abnf S2.1, data-model.md S2.11 note 2). It
// takes no cached-root parameter and has no path that returns a stored value
// without recomputation: T_S's root is never read from
// CommitRingRecord.ledger_root as authoritative. It errors only if the slot
// count exceeds the tree's capacity.
func TSRoot(slots []container.SegmentTableSlot) (Digest, error) {
	return buildTSTree(slots)
}

// LedgerRootVerdict is the outcome of comparing a stored ledger_root against
// a freshly recomputed T_S root.
type LedgerRootVerdict int

const (
	// LedgerRootMatches: the stored ledger_root equals the fresh T_S root.
	LedgerRootMatches LedgerRootVerdict = iota
	// LedgerRootDiverges: the stored ledger_root differs from the fresh T_S
	// root -- a distinct verdict, never silently accepted.
	LedgerRootDiverges
)

// CompareLedgerRoot recomputes T_S fresh from slots and compares it against
// storedLedgerRoot (from CommitRingRecord.ledger_root). It returns
// LedgerRootMatches or LedgerRootDiverges plus the freshly computed root; the
// stored value is used ONLY for comparison, never trusted as the answer
// (data-model.md S2.11 note 2). A capacity error is surfaced.
func CompareLedgerRoot(slots []container.SegmentTableSlot, storedLedgerRoot Digest) (LedgerRootVerdict, Digest, error) {
	fresh, err := TSRoot(slots)
	if err != nil {
		return LedgerRootDiverges, Digest{}, err
	}
	if fresh == storedLedgerRoot {
		return LedgerRootMatches, fresh, nil
	}
	return LedgerRootDiverges, fresh, nil
}
