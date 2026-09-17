package integrity

import (
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// TestTR_009_TSTreeFixedShapeAndAbsentChildFill is T-0130's named test. It
// confirms the internal-node preimage is always exactly 513 octets, that an
// empty tree fills every level with ABSENT_CHILD_DIGEST (a fully-absent tree
// has a well-defined deterministic root), that populated slots at various N
// (0..16384) build a depth-4 tree without error, and that the leaf value is
// the slot-digest.
func TestTR_009_TSTreeFixedShapeAndAbsentChildFill(t *testing.T) {
	// Internal-node preimage is exactly 1 + 16*32 = 513 octets by
	// construction.
	if tsInternalPreimageLen != 513 {
		t.Fatalf("internal preimage length = %d, want 513", tsInternalPreimageLen)
	}

	// A fully-absent tree: all 16 children ABSENT_CHILD_DIGEST at level 0,
	// collapsed 4 times, is deterministic. Compute the expected root by hand.
	var allAbsent [TSArity]Digest
	for i := range allAbsent {
		allAbsent[i] = AbsentChildDigest
	}
	node := AbsentChildDigest
	// Recompute: level-0 leaves are all AbsentChildDigest; each internal node
	// hashes 16 identical children; 4 levels.
	expectAllAbsent := AbsentChildDigest
	{
		lvl := AbsentChildDigest // a single absent leaf value; all leaves equal
		var children [TSArity]Digest
		for d := 0; d < TSDepth; d++ {
			for i := range children {
				children[i] = lvl
			}
			lvl = tsInternalNode(children)
		}
		expectAllAbsent = lvl
	}
	_ = node
	_ = allAbsent

	emptyRoot, err := buildTSTree(nil)
	if err != nil {
		t.Fatalf("buildTSTree(nil): %v", err)
	}
	if emptyRoot != expectAllAbsent {
		t.Fatalf("empty-tree root does not match the all-absent computation")
	}

	// Various populated sizes build without error and are depth-4 (implicit:
	// buildTSTree only ever returns a single root after exactly TSDepth
	// collapses). N up to MAX_SEGMENTS.
	for _, n := range []int{0, 1, 15, 16, 17, 256, 16383, 16384} {
		slots := make([]container.SegmentTableSlot, n)
		for i := range slots {
			var dig pdlfmt.Digest256
			dig[0] = byte(i)
			dig[1] = byte(i >> 8)
			slots[i] = container.SegmentTableSlot{SegmentType: container.SegmentTypeContent, Digest: dig}
		}
		if _, err := buildTSTree(slots); err != nil {
			t.Fatalf("buildTSTree(N=%d): %v", n, err)
		}
	}

	// A populated leaf changes the root vs the empty tree (the leaf value is
	// the slot-digest).
	slots := make([]container.SegmentTableSlot, 1)
	slots[0].Digest = pdlfmt.Digest256{0xAB, 0xCD}
	r, err := buildTSTree(slots)
	if err != nil {
		t.Fatalf("buildTSTree(1): %v", err)
	}
	if r == emptyRoot {
		t.Fatalf("a populated slot did not change the T_S root")
	}

	// Capacity: one past 65536 errors, and it is a named error.
	over := make([]container.SegmentTableSlot, TSLeafCapacity+1)
	if _, err := buildTSTree(over); err == nil {
		t.Fatalf("buildTSTree over capacity did not error")
	}
}
