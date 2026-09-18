package render

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_086_PageDirectoryDigestBinding is T-0243's named unit test (FR-086).
// The PageDirectory is bound to a digest over its complete declared input set
// (geometry, referenced content-unit identities, font digests): recomputing
// over an unchanged input set is identical, and changing any one input class
// changes the digest.
func TestFR_086_PageDirectoryDigestBinding(t *testing.T) {
	fd := func(b byte) pdlfmt.Digest256 {
		var d pdlfmt.Digest256
		d[0] = b
		return d
	}
	base := PageDirectoryInputs{
		Geometry:    0x0A0B0C0D,
		Units:       []pdlfmt.UnitID{pdUnit(0x11), pdUnit(0x22), pdUnit(0x33)},
		FontDigests: []pdlfmt.Digest256{fd(0x01), fd(0x02)},
	}

	// Recomputing over an unchanged input set is identical.
	if base.Digest() != base.Digest() {
		t.Fatal("digest not deterministic over an unchanged input set")
	}

	// Font-digest SET order does not matter (it is a set).
	reordered := base
	reordered.FontDigests = []pdlfmt.Digest256{fd(0x02), fd(0x01)}
	if reordered.Digest() != base.Digest() {
		t.Error("font-digest set order must not change the digest")
	}

	// Changing geometry changes the digest.
	g := base
	g.Geometry = 0xFFFFFFFF
	if g.Digest() == base.Digest() {
		t.Error("changing geometry must change the digest")
	}

	// Changing a referenced unit identity changes the digest.
	u := base
	u.Units = []pdlfmt.UnitID{pdUnit(0x11), pdUnit(0x99), pdUnit(0x33)}
	if u.Digest() == base.Digest() {
		t.Error("changing a referenced unit must change the digest")
	}

	// Adding/removing a referenced unit changes the digest (count is bound).
	u2 := base
	u2.Units = []pdlfmt.UnitID{pdUnit(0x11), pdUnit(0x22)}
	if u2.Digest() == base.Digest() {
		t.Error("changing the referenced-unit count must change the digest")
	}

	// Changing a font digest changes the digest.
	f := base
	f.FontDigests = []pdlfmt.Digest256{fd(0x01), fd(0x77)}
	if f.Digest() == base.Digest() {
		t.Error("changing a font digest must change the digest")
	}

	// Unit ORDER is part of identity (units are consumed in order).
	o := base
	o.Units = []pdlfmt.UnitID{pdUnit(0x33), pdUnit(0x22), pdUnit(0x11)}
	if o.Digest() == base.Digest() {
		t.Error("reordering referenced units must change the digest")
	}
}
