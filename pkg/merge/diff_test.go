package merge

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestTR_002_DiffReportsConstructsNotStorageUnits is T-0239's named unit test
// (TR-002). DiffConstructs expresses differences in terms of document
// constructs (keyed by unit identity), reports NO construct on which the two
// inputs agree, and the reported construct count equals the number actually
// changed -- a one-construct edit is not reported as a whole-file rewrite.
func TestTR_002_DiffReportsConstructsNotStorageUnits(t *testing.T) {
	unchanged1 := mTarget(0x01)
	unchanged2 := mTarget(0x02)
	modified := mTarget(0x03)
	removed := mTarget(0x04)
	added := mTarget(0x05)

	left := map[pdlfmt.UnitID][]byte{
		unchanged1: []byte("alpha"),
		unchanged2: []byte("beta"),
		modified:   []byte("old"),
		removed:    []byte("gone"),
	}
	right := map[pdlfmt.UnitID][]byte{
		unchanged1: []byte("alpha"), // agree
		unchanged2: []byte("beta"),  // agree
		modified:   []byte("new"),   // changed
		added:      []byte("fresh"), // added
	}

	changes := DiffConstructs(left, right)

	// Exactly the 3 constructs that changed (modified, removed, added); the 2
	// agreed constructs are not reported.
	if len(changes) != 3 {
		t.Fatalf("reported %d constructs, want 3 (the number actually changed); changes=%+v", len(changes), changes)
	}

	byID := map[pdlfmt.UnitID]ConstructChange{}
	for _, c := range changes {
		byID[c.Construct] = c
	}
	if _, ok := byID[unchanged1]; ok {
		t.Error("a construct the inputs agree on must not be reported (TR-002)")
	}
	if _, ok := byID[unchanged2]; ok {
		t.Error("a construct the inputs agree on must not be reported (TR-002)")
	}
	if byID[modified].Kind != Modified || string(byID[modified].Left) != "old" || string(byID[modified].Right) != "new" {
		t.Errorf("modified construct wrong: %+v", byID[modified])
	}
	if byID[removed].Kind != Removed {
		t.Errorf("removed construct kind = %v, want Removed", byID[removed].Kind)
	}
	if byID[added].Kind != Added {
		t.Errorf("added construct kind = %v, want Added", byID[added].Kind)
	}

	// Every reported change is keyed by a construct identity, never a storage
	// unit -- the Construct field is a unit id, and there is no segment/octet
	// field to report a storage unit through.
	for _, c := range changes {
		if c.Construct == (pdlfmt.UnitID{}) {
			t.Error("a change must be keyed by a real construct identity")
		}
	}
}
