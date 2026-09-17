package extract_test

import (
	"testing"

	"Protodoc/pkg/extract"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_046_OmitsPageAdjunctAndReportsStaleWhenPaginationAbsentOrStale is
// T-0093's named conformance test. A digest-mismatched (stale) PageDirectory
// and an absent one both yield nil page adjuncts on every locator and
// PaginationStale=true, rather than a guessed page number (FR-046).
func TestFR_046_OmitsPageAdjunctAndReportsStaleWhenPaginationAbsentOrStale(t *testing.T) {
	mkID := func(b byte) pdlfmt.UnitID {
		var id pdlfmt.UnitID
		id[0] = b
		return id
	}
	u1, u2 := mkID(1), mkID(2)
	locs := []extract.Locator{{UnitID: u1}, {UnitID: u2, ScalarOffset: 3}}

	var currentDigest pdlfmt.Digest256
	currentDigest[0] = 0xAB

	assertAllStale := func(name string, res extract.PaginatedResult) {
		if !res.PaginationStale {
			t.Fatalf("%s: PaginationStale = false, want true", name)
		}
		if len(res.Locators) != len(locs) {
			t.Fatalf("%s: got %d locators, want %d", name, len(res.Locators), len(locs))
		}
		for i, pl := range res.Locators {
			if pl.Page != nil {
				t.Fatalf("%s: locator %d has a non-nil page adjunct %d (must be omitted, no guessing)", name, i, pl.Page.Page)
			}
		}
	}

	// Stale: directory's input digest does not match current content.
	var staleDigest pdlfmt.Digest256
	staleDigest[0] = 0x99 // different from currentDigest
	staleDir := &extract.PageDirectory{
		InputDigest: staleDigest,
		Pages:       map[pdlfmt.UnitID]uint32{u1: 1, u2: 2},
	}
	assertAllStale("stale directory", extract.PaginateLocators(locs, staleDir, currentDigest))

	// Absent: no directory at all.
	assertAllStale("absent directory", extract.PaginateLocators(locs, nil, currentDigest))
}
