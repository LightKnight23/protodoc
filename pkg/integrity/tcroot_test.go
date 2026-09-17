package integrity

import (
	"reflect"
	"testing"
)

// TestFR_001_TCRootIsCompleteStateValueSet is T-0136's named test. TCRoot
// over a fixed record set is reproducible; changing any one record's stored
// frame bytes, adding a record, or removing a record all change TCRoot; and
// TCRoot's signature depends only on []ContentRecord (no SegmentTableSlot or
// T_S value).
func TestFR_001_TCRootIsCompleteStateValueSet(t *testing.T) {
	// TCRoot takes only []ContentRecord -- no SegmentTableSlot / T_S input.
	ft := reflect.TypeOf(TCRoot)
	if ft.NumIn() != 1 {
		t.Fatalf("TCRoot takes %d params, want 1", ft.NumIn())
	}

	base := []ContentRecord{
		mkRec(1, false),
		mkRec(2, true),
		mkRec(3, false),
	}

	r1, err := TCRoot(base)
	if err != nil {
		t.Fatalf("TCRoot: %v", err)
	}
	// Reproducible.
	r1b, _ := TCRoot(base)
	if r1 != r1b {
		t.Fatalf("TCRoot not reproducible over the identical record set")
	}

	// Changing one record's frame changes the root.
	changed := append([]ContentRecord(nil), base...)
	changed[1].Frame = append([]byte(nil), base[1].Frame...)
	changed[1].Frame[1] ^= 0xFF
	if rc, _ := TCRoot(changed); rc == r1 {
		t.Fatalf("changing a record frame did not change TCRoot")
	}

	// Changing a redactable record's salt changes the root.
	saltChanged := append([]ContentRecord(nil), base...)
	saltChanged[1].Salt[0] ^= 0xFF
	if rs, _ := TCRoot(saltChanged); rs == r1 {
		t.Fatalf("changing a redactable salt did not change TCRoot")
	}

	// Adding a record changes the root.
	added := append(append([]ContentRecord(nil), base...), mkRec(4, false))
	if ra, _ := TCRoot(added); ra == r1 {
		t.Fatalf("adding a record did not change TCRoot")
	}

	// Removing a record changes the root.
	removed := append([]ContentRecord(nil), base[:2]...)
	if rr, _ := TCRoot(removed); rr == r1 {
		t.Fatalf("removing a record did not change TCRoot")
	}

	// Order-independence of the INPUT (traversal orders internally): the
	// same set in a different input order yields the same root.
	shuffled := []ContentRecord{base[2], base[0], base[1]}
	if rsh, _ := TCRoot(shuffled); rsh != r1 {
		t.Fatalf("TCRoot depends on input order; it must order internally")
	}
}
