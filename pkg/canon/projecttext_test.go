package canon

import (
	"strings"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestTR_004_TextProjectionIsDeterministic is T-0319's named unit test
// (TR-004). ProjectText is a deterministic, human-readable projection of C(S),
// driven by the canonical traversal: identical state (and any storage
// permutation of it) yields identical text across runs.
func TestTR_004_TextProjectionIsDeterministic(t *testing.T) {
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

	t1 := ProjectText(state)
	t2 := ProjectText(state)
	if t1 != t2 {
		t.Errorf("text projection not deterministic across runs")
	}

	// Storage permutation yields identical projection (canonical order).
	shuffled := &Document{Subtrees: []ContentSubtree{subs[1], subs[2], subs[0]}}
	if ProjectText(shuffled) != t1 {
		t.Errorf("text projection depends on storage order")
	}

	// Canonical order: 0x10 line precedes 0x20 precedes 0x30.
	lines := strings.Split(strings.TrimRight(t1, "\n"), "\n")
	if len(lines) != 3 || !strings.HasPrefix(lines[0], "10") || !strings.HasPrefix(lines[1], "20") || !strings.HasPrefix(lines[2], "30") {
		t.Errorf("text projection not in canonical order: %q", t1)
	}

	// BOTTOM state projects to empty text.
	if ProjectText(&Document{}) != "" {
		t.Errorf("BOTTOM state must project to empty text")
	}
}
