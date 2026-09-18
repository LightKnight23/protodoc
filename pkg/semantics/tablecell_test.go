package semantics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConformanceTBL000ScopeRoundtrip is T-0281's named conformance test
// (vector CONFORMANCE-TBL-000-scope-roundtrip; FR-039). The header-cell scope
// field round-trips through encode/decode for every scope value, reserved
// values are rejected, and document.abnf/data-model.md declare the field.
func TestConformanceTBL000ScopeRoundtrip(t *testing.T) {
	scopes := []CellScope{ScopeRow, ScopeColumn, ScopeRowGroup, ScopeColumnGroup}
	for _, sc := range scopes {
		enc, err := EncodeCellKindScope(nil, CellHeader, sc)
		if err != nil {
			t.Fatalf("encode scope %d: %v", sc, err)
		}
		k, s, n, err := DecodeCellKindScope(enc)
		if err != nil {
			t.Fatalf("decode scope %d: %v", sc, err)
		}
		if k != CellHeader || s != sc || n != 2 {
			t.Errorf("roundtrip scope %d: got kind=%d scope=%d n=%d", sc, k, s, n)
		}
	}

	// Reserved kind (0x02) and scope (0x04) are rejected.
	if _, err := EncodeCellKindScope(nil, CellKind(0x02), ScopeRow); err == nil {
		t.Errorf("reserved cell-kind 0x02 must be rejected on encode")
	}
	if _, err := EncodeCellKindScope(nil, CellHeader, CellScope(0x04)); err == nil {
		t.Errorf("reserved cell-scope 0x04 must be rejected on encode")
	}
	if _, _, _, err := DecodeCellKindScope([]byte{0x01, 0x04}); err == nil {
		t.Errorf("reserved cell-scope 0x04 must be rejected on decode")
	}

	// The wire grammar and the entity table declare the fields.
	abnf, err := os.ReadFile(filepath.Join("..", "..", "specs", "001-protodoc-format-core", "contracts", "document.abnf"))
	if err != nil {
		t.Fatalf("read document.abnf: %v", err)
	}
	for _, tok := range []string{"cell-kind", "cell-scope"} {
		if !strings.Contains(string(abnf), tok) {
			t.Errorf("document.abnf must define %s", tok)
		}
	}
	dm, err := os.ReadFile(filepath.Join("..", "..", "specs", "001-protodoc-format-core", "data-model.md"))
	if err != nil {
		t.Fatalf("read data-model.md: %v", err)
	}
	for _, tok := range []string{"cell_kind", "cell_scope"} {
		if !strings.Contains(string(dm), tok) {
			t.Errorf("data-model.md must document %s", tok)
		}
	}
}
