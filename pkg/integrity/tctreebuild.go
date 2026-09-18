// T_C tree assembly (T-0135, FR-003; integrity.abnf S2.2): the arity-16,
// depth<=5 content-commitment tree. Leaves are the redactable (0x02) or
// non-redactable (0x07) leaf digests of each content-model record, placed in
// T_C subtree ordinal order (T-0134). Internal nodes are 0x08 || 16 child
// digests (always a 513-octet preimage). Every position with no
// corresponding subtree -- at every level -- is ABSENT_CHILD_DIGEST.
package integrity

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// T_C fixed-shape constants (integrity.abnf S2.2).
const (
	// TCArity is the fixed number of children at every T_C internal node.
	TCArity = 16
	// TCMaxDepth is the maximum tree depth (16^5 = 1048576 = MAX_CONTENT_UNITS).
	TCMaxDepth = 5
	// TCMaxLeafCapacity is 16^5 = 1048576.
	TCMaxLeafCapacity = 1048576
	// tcInternalPreimageLen is 1 + 16*32 = 513 octets.
	tcInternalPreimageLen = 1 + TCArity*32
)

// ErrTCTooManyRecords is returned when more content records are supplied
// than the depth-5 tree can hold (its 1048576-leaf capacity).
var ErrTCTooManyRecords = errors.New("integrity: more content records than the T_C depth-5 capacity (1048576)")

// tcInternalNode hashes an internal node: 0x08 || 16 child digests, always
// exactly 513 octets.
func tcInternalNode(children [TCArity]Digest) Digest {
	var buf [tcInternalPreimageLen]byte
	buf[0] = DomainTCInternal
	for i := 0; i < TCArity; i++ {
		copy(buf[1+i*32:1+(i+1)*32], children[i][:])
	}
	var d Digest
	sum := sha256.Sum256(buf[:])
	copy(d[:], sum[:])
	return d
}

// leafDigest computes the leaf digest for one record: redactable (0x02 with
// salt) or non-redactable (0x07) per its designation.
func leafDigest(rec ContentRecord) Digest {
	if rec.Redacted {
		// A redacted redactable record's leaf is its retained commitment
		// digest (frame and salt gone); this equals the pre-redaction
		// redactable-leaf digest, so the T_C root is preserved (FR-076).
		return rec.RetainedLeaf
	}
	if rec.Redactable {
		return TCLeafRedactable(rec.Salt, rec.Frame)
	}
	return TCLeafNonredactable(rec.Frame)
}

// tcDepthFor returns the minimal T_C depth d in [1, TCMaxDepth] whose leaf
// capacity 16^d is >= n (n >= 1). n == 0 uses depth 1 (a single full node of
// 16 absent children). It is a deterministic function of n, so two
// implementations build the identical tree shape for the same record count.
func tcDepthFor(n int) (int, int, error) {
	if n > TCMaxLeafCapacity {
		return 0, 0, fmt.Errorf("%w: got %d", ErrTCTooManyRecords, n)
	}
	depth := 1
	capacity := TCArity
	for capacity < n {
		depth++
		capacity *= TCArity
	}
	return depth, capacity, nil
}

// buildTCTree builds the T_C tree over records (already understood to be in
// arbitrary order; this function orders them by the interim traversal rule)
// and returns the root. The tree depth is the minimal depth (<= 5) holding
// all records; every absent leaf/child position at every level is
// ABSENT_CHILD_DIGEST.
func buildTCTree(records []ContentRecord) (Digest, error) {
	return buildTCTreeOrdered(OrderRecords(records))
}

// buildTCTreeOrdered builds the fixed-arity T_C tree over an already-ordered
// record slice and returns the root.
func buildTCTreeOrdered(ordered []ContentRecord) (Digest, error) {
	depth, capacity, err := tcDepthFor(len(ordered))
	if err != nil {
		return Digest{}, err
	}

	// Level 0: capacity leaves; present positions take the leaf digest,
	// absent positions take ABSENT_CHILD_DIGEST.
	level := make([]Digest, capacity)
	for i := range level {
		if i < len(ordered) {
			level[i] = leafDigest(ordered[i])
		} else {
			level[i] = AbsentChildDigest
		}
	}

	// Collapse `depth` times.
	for d := 0; d < depth; d++ {
		next := make([]Digest, len(level)/TCArity)
		for p := range next {
			var children [TCArity]Digest
			copy(children[:], level[p*TCArity:(p+1)*TCArity])
			next[p] = tcInternalNode(children)
		}
		level = next
	}
	if len(level) != 1 {
		return Digest{}, fmt.Errorf("integrity: T_C collapse produced %d roots, want 1", len(level))
	}
	return level[0], nil
}

// TCRoot returns the T_C root -- the document state identity (DP-006/FR-003:
// "every state gets an identifier differing whenever any value of that state
// differs"), and the concrete realization of FR-001's "document state =
// complete value set". Its input `records` IS the value set that determines
// extraction, render, verdict and metadata (ATTEST-typed segments are
// excluded upstream; see T-0138). It is recomputed fresh from the records'
// stored frame bytes and depends on no SegmentTableSlot and no T_S value:
// T_C_root is signed directly inside signed_object, never routed through
// T_S (integrity.abnf S2.2, data-model.md S2.10). It errors only if the
// record count exceeds the T_C capacity.
func TCRoot(records []ContentRecord) (Digest, error) {
	return buildTCTree(records)
}

// TCRootWithSequence returns the T_C root with the subtree traversal keyed on
// the document's authored ROOT_SEQUENCE reading order (`order`, the rs-order
// list) instead of the FLAGGED CSPRNG interim rule (FR-036, T-0275). Because
// the ordering is a pure function of ROOT_SEQUENCE, the returned root changes
// when — and only when — ROOT_SEQUENCE changes: permuting the storage order of
// `records` yields the identical root, while reordering `order` yields a
// different root. It errors only if the record count exceeds T_C capacity.
func TCRootWithSequence(records []ContentRecord, order []pdlfmt.UnitID) (Digest, error) {
	return buildTCTreeOrdered(OrderRecordsBySequence(records, order))
}
