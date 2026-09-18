package canon

import (
	"bytes"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_124_EmptyStateEmitsWellDefinedMinimalPrefix is T-0315's named unit
// test (FR-124). The BOTTOM state is the zero-subtree Document; its L* emission
// is a valid, minimal, well-defined file: exactly the fixed prefix and zero
// content bytes.
func TestFR_124_EmptyStateEmitsWellDefinedMinimalPrefix(t *testing.T) {
	empty := EmptyState()
	if !IsBottom(empty) {
		t.Errorf("EmptyState must be recognised as BOTTOM")
	}
	if !IsBottom(nil) {
		t.Errorf("nil state must be recognised as BOTTOM")
	}

	// A non-empty state is not BOTTOM.
	var id pdlfmt.UnitID
	id[0] = 0x1
	if IsBottom(&Document{Subtrees: []ContentSubtree{{UnitID: id}}}) {
		t.Errorf("a state with content must not be BOTTOM")
	}

	// EmitBottom yields exactly the fixed prefix and no content.
	f := EmitBottom()
	if len(f) != MinimalPrefixLen {
		t.Fatalf("EmitBottom length = %d, want %d (fixed prefix only)", len(f), MinimalPrefixLen)
	}

	// C(BOTTOM) is empty (zero content octets stream past the prefix).
	var canon bytes.Buffer
	if err := Canonicalize(empty, &canon); err != nil {
		t.Fatalf("Canonicalize(BOTTOM): %v", err)
	}
	if canon.Len() != 0 {
		t.Errorf("C(BOTTOM) must be zero content bytes, got %d", canon.Len())
	}
}
