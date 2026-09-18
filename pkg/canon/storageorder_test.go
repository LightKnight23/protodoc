package canon

import (
	"bytes"
	"math/rand"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestCONF_CANON_001_StorageOrderIndependence is T-0309's named conformance
// test (vector CONF-CANON-001-storage-order-independence; NFR-002). A corpus of
// >=10 fixture states, each canonicalized from >=2 distinct physical
// storage-order layouts, must yield byte-identical C(S) for every layout — C(S)
// is a pure function of the logical state, not its storage order.
func TestCONF_CANON_001_StorageOrderIndependence(t *testing.T) {
	const fixtures = 12
	r := rand.New(rand.NewSource(20260918))

	for f := 0; f < fixtures; f++ {
		// Build a logical state with a fixture-dependent number of subtrees.
		n := 3 + f
		base := make([]ContentSubtree, n)
		for i := 0; i < n; i++ {
			var id pdlfmt.UnitID
			id[0] = byte((i*37 + f*11) & 0xFF)
			id[1] = byte(i)
			id[15] = byte(f)
			base[i] = ContentSubtree{UnitID: id, Frame: []byte{byte(i), byte(f), 0xAB}}
		}

		// Layout A: the base storage order.
		var canonA bytes.Buffer
		if err := Canonicalize(&Document{Subtrees: append([]ContentSubtree(nil), base...)}, &canonA); err != nil {
			t.Fatalf("fixture %d layout A: %v", f, err)
		}

		// Layouts B..D: distinct shuffles of the same logical state.
		for layout := 0; layout < 3; layout++ {
			shuffled := append([]ContentSubtree(nil), base...)
			r.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
			var canonB bytes.Buffer
			if err := Canonicalize(&Document{Subtrees: shuffled}, &canonB); err != nil {
				t.Fatalf("fixture %d layout %d: %v", f, layout, err)
			}
			if !bytes.Equal(canonA.Bytes(), canonB.Bytes()) {
				t.Errorf("fixture %d: C(S) differs between storage layouts (layout %d)", f, layout)
			}
		}
	}
}
