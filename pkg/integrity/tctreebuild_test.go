package integrity

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

func mkRec(seed int, redactable bool) ContentRecord {
	var id pdlfmt.UnitID
	id[0] = byte(seed * 13)
	for k := 0; k < 8; k++ {
		id[8+k] = byte(uint64(seed) >> (uint(k) * 8))
	}
	var salt [SaltSize]byte
	salt[0] = byte(seed)
	return ContentRecord{UnitID: id, Frame: []byte{0x01, byte(seed)}, Redactable: redactable, Salt: salt}
}

// TestFR_003_TCTreeFixedShapeAndAbsentChildFill is T-0135's named test.
// Given a mixed set of redactable and non-redactable records, the builder
// produces a tree whose depth never exceeds 5, every internal-node preimage
// is exactly 513 octets, and absent positions at every level equal
// ABSENT_CHILD_DIGEST.
func TestFR_003_TCTreeFixedShapeAndAbsentChildFill(t *testing.T) {
	if tcInternalPreimageLen != 513 {
		t.Fatalf("T_C internal preimage length = %d, want 513", tcInternalPreimageLen)
	}

	// Depth is minimal and never exceeds 5.
	depthCases := []struct {
		n         int
		wantDepth int
	}{
		{0, 1}, {1, 1}, {16, 1}, {17, 2}, {256, 2}, {257, 3}, {4096, 3}, {65536, 4}, {65537, 5}, {1048576, 5},
	}
	for _, c := range depthCases {
		d, _, err := tcDepthFor(c.n)
		if err != nil {
			t.Fatalf("tcDepthFor(%d): %v", c.n, err)
		}
		if d != c.wantDepth {
			t.Fatalf("tcDepthFor(%d) = %d, want %d", c.n, d, c.wantDepth)
		}
		if d > TCMaxDepth {
			t.Fatalf("tcDepthFor(%d) = %d exceeds TCMaxDepth", c.n, d)
		}
	}

	// One past capacity errors.
	if _, _, err := tcDepthFor(TCMaxLeafCapacity + 1); err == nil {
		t.Fatalf("tcDepthFor over capacity did not error")
	}

	// An empty tree fills all leaves with ABSENT_CHILD_DIGEST -> a
	// well-defined deterministic root: one 0x08 node over 16 absent children.
	var absentChildren [TCArity]Digest
	for i := range absentChildren {
		absentChildren[i] = AbsentChildDigest
	}
	wantEmpty := tcInternalNode(absentChildren)
	emptyRoot, err := buildTCTree(nil)
	if err != nil {
		t.Fatalf("buildTCTree(nil): %v", err)
	}
	if emptyRoot != wantEmpty {
		t.Fatalf("empty T_C root != single all-absent node")
	}

	// A mixed set of redactable and non-redactable records builds without
	// error at several sizes.
	for _, n := range []int{1, 5, 16, 17, 100} {
		recs := make([]ContentRecord, n)
		for i := range recs {
			recs[i] = mkRec(i, i%2 == 0)
		}
		if _, err := buildTCTree(recs); err != nil {
			t.Fatalf("buildTCTree(n=%d): %v", n, err)
		}
	}

	// A redactable record and a non-redactable record with identical frames
	// produce different roots (leaf domain separation flows into the root).
	rRed := mkRec(1, true)
	rNon := mkRec(1, false) // same seed => same id/frame, differs only in Redactable
	rootRed, _ := buildTCTree([]ContentRecord{rRed})
	rootNon, _ := buildTCTree([]ContentRecord{rNon})
	if rootRed == rootNon {
		t.Fatalf("redactable vs non-redactable single-record roots collided")
	}
}
