package canon

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestNFR_004_FullCompactionRefusedWhenSignaturePresent is T-0310's named unit
// test (NFR-004). Compact refuses (with a named error, before any I/O) when a
// Signature record is present, and otherwise produces L*(S) = C(S).
func TestNFR_004_FullCompactionRefusedWhenSignaturePresent(t *testing.T) {
	var id pdlfmt.UnitID
	id[0] = 0x11
	doc := &Document{Subtrees: []ContentSubtree{{UnitID: id, Frame: []byte("hello")}}}

	// Signature present: refused with the named error, no file produced.
	f, err := Compact(CompactInput{State: doc, SignaturePresent: true})
	if !errors.Is(err, ErrCompactionRefusedSignaturePresent) {
		t.Fatalf("with signature: err = %v, want ErrCompactionRefusedSignaturePresent", err)
	}
	if f != nil {
		t.Errorf("refusal must produce no file, got %+v", f)
	}

	// No signature: compaction proceeds and equals C(S).
	f, err = Compact(CompactInput{State: doc, SignaturePresent: false})
	if err != nil {
		t.Fatalf("without signature: %v", err)
	}
	if f == nil || string(f.Bytes) != "hello" {
		t.Errorf("compacted bytes = %q, want C(S) = %q", string(f.Bytes), "hello")
	}
}
