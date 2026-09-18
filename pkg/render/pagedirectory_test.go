package render

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// pdUnit builds a deterministic content-unit identity from a seed byte.
func pdUnit(seed byte) pdlfmt.UnitID {
	var id pdlfmt.UnitID
	id[0] = seed
	return id
}

// TestNFR_010_PageDirectoryKeyedByContentIdentity is T-0242's named unit test
// (NFR-010). The PageDirectory keys entries by content-unit identity, deriving
// each page's ordinal as the entry's rank; a single-page insert or remove
// touches only the affected entry (plus its rank path), never renumbering the
// other entries' stored keys.
func TestNFR_010_PageDirectoryKeyedByContentIdentity(t *testing.T) {
	// Ordinals are DERIVED from rank, never stored: build a directory and
	// confirm ranks reflect identity order.
	seeds := []byte{0x10, 0x20, 0x30, 0x40, 0x50}
	build := func() *PageDirectory {
		d := &PageDirectory{}
		for _, s := range seeds {
			d.Insert(PageEntry{BreakUnit: pdUnit(s), Geometry: uint32(s)})
		}
		return d
	}

	tests := []struct {
		name    string
		mutate  func(d *PageDirectory)
		wantKey pdlfmt.UnitID
	}{
		{"insert-at-start", func(d *PageDirectory) { d.Insert(PageEntry{BreakUnit: pdUnit(0x05)}) }, pdUnit(0x05)},
		{"insert-in-middle", func(d *PageDirectory) { d.Insert(PageEntry{BreakUnit: pdUnit(0x35)}) }, pdUnit(0x35)},
		{"insert-at-end", func(d *PageDirectory) { d.Insert(PageEntry{BreakUnit: pdUnit(0x60)}) }, pdUnit(0x60)},
		{"remove-at-start", func(d *PageDirectory) { d.Remove(pdUnit(0x10)) }, pdUnit(0x10)},
		{"remove-in-middle", func(d *PageDirectory) { d.Remove(pdUnit(0x30)) }, pdUnit(0x30)},
		{"remove-at-end", func(d *PageDirectory) { d.Remove(pdUnit(0x50)) }, pdUnit(0x50)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := build()
			before := d.Entries()
			tc.mutate(d)

			// Only the affected entry is reported as touched.
			touched := d.Touched()
			if len(touched) != 1 || touched[0] != tc.wantKey {
				t.Fatalf("%s: touched %v, want exactly [%x]", tc.name, touched, tc.wantKey)
			}

			// Every OTHER entry that survived keeps its identity key unchanged
			// (no renumbering of stored keys -- ordinals shift by rank only).
			after := d.Entries()
			survivors := map[pdlfmt.UnitID]bool{}
			for _, e := range after {
				survivors[e.BreakUnit] = true
			}
			for _, e := range before {
				if e.BreakUnit == tc.wantKey {
					continue // the mutated entry
				}
				if !survivors[e.BreakUnit] {
					t.Errorf("%s: pre-existing entry %x lost", tc.name, e.BreakUnit)
				}
			}
		})
	}

	// Single-entry corpus: insert then remove the only entry.
	d := &PageDirectory{}
	d.Insert(PageEntry{BreakUnit: pdUnit(0x01)})
	if r, ok := d.Rank(pdUnit(0x01)); !ok || r != 0 {
		t.Fatalf("single-entry rank = %d,%v, want 0,true", r, ok)
	}
	d.Remove(pdUnit(0x01))
	if d.Len() != 0 {
		t.Fatalf("single-entry remove left %d entries", d.Len())
	}

	// Ranks derive ordinals in identity order after edits.
	d = build()
	for wantRank, s := range seeds {
		if r, ok := d.Rank(pdUnit(s)); !ok || r != wantRank {
			t.Errorf("rank(%x) = %d,%v, want %d,true", pdUnit(s), r, ok, wantRank)
		}
	}
}
