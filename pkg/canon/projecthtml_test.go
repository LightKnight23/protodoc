package canon

import (
	"strings"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestTR_004_HtmlProjectionIsDeterministic is T-0320's named unit test
// (TR-004). ProjectHTML is a deterministic markup projection of C(S) sharing
// the canonical traversal with the text projector: identical state (and any
// storage permutation) yields identical HTML across runs.
func TestTR_004_HtmlProjectionIsDeterministic(t *testing.T) {
	u := func(b byte) pdlfmt.UnitID {
		var id pdlfmt.UnitID
		id[0] = b
		return id
	}
	subs := []ContentSubtree{
		{UnitID: u(0x30), Frame: []byte{0xAA}},
		{UnitID: u(0x10), Frame: []byte{0xBB}},
		{UnitID: u(0x20), Frame: []byte{0xCC}},
	}
	state := &Document{Subtrees: subs}

	h1 := ProjectHTML(state)
	if ProjectHTML(state) != h1 {
		t.Errorf("html projection not deterministic across runs")
	}
	shuffled := &Document{Subtrees: []ContentSubtree{subs[2], subs[0], subs[1]}}
	if ProjectHTML(shuffled) != h1 {
		t.Errorf("html projection depends on storage order")
	}

	// Canonical order: 0x10 div precedes 0x20 precedes 0x30.
	i10 := strings.Index(h1, `data-unit="10`)
	i20 := strings.Index(h1, `data-unit="20`)
	i30 := strings.Index(h1, `data-unit="30`)
	if !(i10 >= 0 && i10 < i20 && i20 < i30) {
		t.Errorf("html divs not in canonical order (10@%d 20@%d 30@%d)", i10, i20, i30)
	}

	// BOTTOM state projects to the empty skeleton (no data-unit divs).
	if strings.Contains(ProjectHTML(&Document{}), "data-unit") {
		t.Errorf("BOTTOM state must have no content divs")
	}
}
