package render

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_088_StalePageDirectoryRefusedAndRederived is T-0244's named
// integration test (FR-087, FR-088). The PageDirectory is non-normative
// (FR-087). On open, a stored directory whose input digest matches current
// inputs is served; a stored directory whose digest is stale (hand-corrupted)
// is discarded and re-derived from the authoritative content before any page
// render, and the render output is identical to a document that never had a
// stale directory (FR-088).
func TestFR_088_StalePageDirectoryRefusedAndRederived(t *testing.T) {
	// FR-087: the PageDirectory is non-normative.
	if (*PageDirectory)(nil).Normativity() != NonNormative {
		t.Fatal("PageDirectory must be non-normative relative to content (FR-087)")
	}

	inputs := PageDirectoryInputs{
		Geometry: 0x11223344,
		Units:    []pdlfmt.UnitID{pdUnit(0x10), pdUnit(0x20), pdUnit(0x30)},
	}

	// The authoritative re-derivation: builds the directory from the units.
	deriveCalled := 0
	derive := func(in PageDirectoryInputs) *PageDirectory {
		deriveCalled++
		d := &PageDirectory{}
		for _, u := range in.Units {
			d.Insert(PageEntry{BreakUnit: u, Geometry: in.Geometry})
		}
		return d
	}

	// Reference: a document that never had a stale directory (freshly derived).
	reference := derive(inputs)
	refEntries := reference.Entries()

	// Current stored directory: digest matches -> served as-is, no re-derive.
	freshStored := &StoredPageDirectory{
		Entries:      refEntries,
		StoredDigest: inputs.Digest(),
	}
	before := deriveCalled
	served, rederived := OpenPageDirectory(freshStored, inputs, derive)
	if rederived {
		t.Error("current directory should be served, not re-derived")
	}
	if deriveCalled != before {
		t.Error("current directory must not trigger derivation")
	}
	if len(served.Entries()) != len(refEntries) {
		t.Error("served current directory differs from reference")
	}

	// Stale stored directory: hand-corrupt the digest -> refused + re-derived.
	var corrupt pdlfmt.Digest256
	corrupt[0] = 0xDE
	corrupt[1] = 0xAD
	staleStored := &StoredPageDirectory{
		Entries:      []PageEntry{{BreakUnit: pdUnit(0xFF), Geometry: 0}}, // wrong/garbage content
		StoredDigest: corrupt,
	}
	before = deriveCalled
	got, rederived := OpenPageDirectory(staleStored, inputs, derive)
	if !rederived {
		t.Fatal("stale directory must be refused and re-derived (FR-088)")
	}
	if deriveCalled != before+1 {
		t.Fatalf("stale directory must trigger exactly one re-derivation, got %d", deriveCalled-before)
	}

	// The re-derived output is IDENTICAL to the never-stale reference: the
	// garbage stored entries were discarded, not served.
	gotEntries := got.Entries()
	if len(gotEntries) != len(refEntries) {
		t.Fatalf("re-derived %d entries, want %d (reference)", len(gotEntries), len(refEntries))
	}
	for i := range gotEntries {
		if gotEntries[i].BreakUnit != refEntries[i].BreakUnit || gotEntries[i].Geometry != refEntries[i].Geometry {
			t.Errorf("re-derived entry %d = %+v, want %+v (reference)", i, gotEntries[i], refEntries[i])
		}
	}

	// Absent stored directory (nil) is also re-derived before any render.
	before = deriveCalled
	if _, rederived := OpenPageDirectory(nil, inputs, derive); !rederived || deriveCalled != before+1 {
		t.Error("absent directory must be re-derived")
	}
}
