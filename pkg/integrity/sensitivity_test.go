package integrity

import (
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_003_RootChangesForEveryCoveredValueMutation is T-0143's named
// conformance test. For a base fixture it generates one mutated variant per
// covered leaf/domain-tag kind and asserts the corresponding root changes
// for every variant, while an unmutated control leaves the root unchanged
// (ruling out a test that trivially always reports "changed").
func TestFR_003_RootChangesForEveryCoveredValueMutation(t *testing.T) {
	// --- T_S sensitivity: a leaf slot-digest change changes TSRoot. ---
	baseSlots := make([]container.SegmentTableSlot, 5)
	for i := range baseSlots {
		var d pdlfmt.Digest256
		d[0] = byte(i + 1)
		baseSlots[i] = container.SegmentTableSlot{SegmentType: container.SegmentTypeContent, Digest: d}
	}
	tsBase, _ := TSRoot(baseSlots)

	// Mutation 1: T_S leaf digest.
	tsMut := append([]container.SegmentTableSlot(nil), baseSlots...)
	tsMut[2].Digest[0] ^= 0xFF
	if r, _ := TSRoot(tsMut); r == tsBase {
		t.Fatalf("T_S root unchanged after a leaf slot-digest mutation")
	}
	// Control: identical copy -> identical root.
	tsCtl := append([]container.SegmentTableSlot(nil), baseSlots...)
	if r, _ := TSRoot(tsCtl); r != tsBase {
		t.Fatalf("T_S control (identical copy) changed the root")
	}

	// --- T_C sensitivity: build a base with a redactable and a
	// non-redactable record. ---
	baseRecs := []ContentRecord{
		mkRec(1, true),  // redactable
		mkRec(2, false), // non-redactable
		mkRec(3, true),  // redactable
	}
	tcBase, _ := TCRoot(baseRecs)

	mutate := func(name string, fn func([]ContentRecord)) {
		recs := deepCopyRecs(baseRecs)
		fn(recs)
		if r, _ := TCRoot(recs); r == tcBase {
			t.Fatalf("T_C root unchanged after mutation: %s", name)
		}
	}

	// Mutation 2: redactable leaf frame byte.
	mutate("redactable frame byte", func(r []ContentRecord) { r[0].Frame[1] ^= 0xFF })
	// Mutation 3: redactable leaf salt.
	mutate("redactable salt", func(r []ContentRecord) { r[0].Salt[0] ^= 0xFF })
	// Mutation 4: non-redactable leaf frame byte.
	mutate("nonredactable frame byte", func(r []ContentRecord) { r[1].Frame[1] ^= 0xFF })
	// Mutation 5: removing a record makes a tree position become ABSENT,
	// changing the root.
	removed := deepCopyRecs(baseRecs)[:2]
	if r, _ := TCRoot(removed); r == tcBase {
		t.Fatalf("T_C root unchanged after removing a record")
	}

	// Control: identical deep copy -> identical root.
	if r, _ := TCRoot(deepCopyRecs(baseRecs)); r != tcBase {
		t.Fatalf("T_C control (identical copy) changed the root")
	}
}

func deepCopyRecs(in []ContentRecord) []ContentRecord {
	out := make([]ContentRecord, len(in))
	for i, r := range in {
		out[i] = r
		out[i].Frame = append([]byte(nil), r.Frame...)
	}
	return out
}
