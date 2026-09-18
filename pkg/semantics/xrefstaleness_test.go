package semantics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestConformanceXREF001StalenessWithoutLayout is T-0289's named conformance
// test (vector CONFORMANCE-XREF-001-staleness-without-layout; FR-085). A
// cross-reference binds a digest to its target's state; staleness is reported
// from the digest mismatch alone, without computing any layout.
func TestConformanceXREF001StalenessWithoutLayout(t *testing.T) {
	tgt := pdUnitSem(0x40)
	var d0, d1 pdlfmt.Digest256
	d0[0] = 0xAA
	d1[0] = 0xBB

	ref := BoundCrossReference{XrefID: pdUnitSem(0x01), Target: tgt, BoundDigest: d0}

	// Current: target's current digest matches the bound one -> not stale.
	current := map[pdlfmt.UnitID]pdlfmt.Digest256{tgt: d0}
	if f := CheckXrefStaleness([]BoundCrossReference{ref}, current); len(f) != 0 {
		t.Errorf("matching digest should not be stale, got %+v", f)
	}

	// Target edited: current digest differs -> stale, reported naming the ref.
	edited := map[pdlfmt.UnitID]pdlfmt.Digest256{tgt: d1}
	f := CheckXrefStaleness([]BoundCrossReference{ref}, edited)
	if len(f) != 1 || f[0].XrefID != ref.XrefID || f[0].Target != tgt {
		t.Fatalf("edited target: findings = %+v, want one stale naming the ref/target", f)
	}
	if f[0].BoundDigest != d0 || f[0].CurrentDigest != d1 {
		t.Errorf("finding should report bound %x vs current %x", f[0].BoundDigest[0], f[0].CurrentDigest[0])
	}

	// Absent target -> stale.
	if f := CheckXrefStaleness([]BoundCrossReference{ref}, map[pdlfmt.UnitID]pdlfmt.Digest256{}); len(f) != 1 {
		t.Errorf("absent target should be stale, got %+v", f)
	}

	// The wire grammar declares the digest-binding field.
	abnf, err := os.ReadFile(filepath.Join("..", "..", "specs", "001-protodoc-format-core", "contracts", "document.abnf"))
	if err != nil {
		t.Fatalf("read document.abnf: %v", err)
	}
	if !strings.Contains(string(abnf), "xref-target-digest") {
		t.Errorf("document.abnf must define xref-target-digest")
	}
}
