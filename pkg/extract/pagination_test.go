package extract_test

import (
	"testing"

	"Protodoc/pkg/extract"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_045_EmitsPageAdjunctWhenPaginationCurrent is T-0092's named
// conformance test. A fresh (current) PageDirectory yields a non-nil page
// adjunct on every locator, matching the golden page map exactly.
func TestFR_045_EmitsPageAdjunctWhenPaginationCurrent(t *testing.T) {
	mkID := func(b byte) pdlfmt.UnitID {
		var id pdlfmt.UnitID
		id[0] = b
		return id
	}
	u1, u2, u3 := mkID(1), mkID(2), mkID(3)

	golden := map[pdlfmt.UnitID]uint32{u1: 1, u2: 1, u3: 2}
	var contentDigest pdlfmt.Digest256
	contentDigest[0] = 0xAB

	// A CURRENT directory: its input digest matches the content digest.
	dir := &extract.PageDirectory{InputDigest: contentDigest, Pages: golden}

	locs := []extract.Locator{
		{UnitID: u1, ScalarOffset: 0},
		{UnitID: u2, ScalarOffset: 5},
		{UnitID: u3, ScalarOffset: 0},
	}

	res := extract.PaginateLocators(locs, dir, contentDigest)
	if res.PaginationStale {
		t.Fatalf("current directory reported stale")
	}
	if len(res.Locators) != len(locs) {
		t.Fatalf("got %d paginated locators, want %d", len(res.Locators), len(locs))
	}
	for i, pl := range res.Locators {
		if pl.Page == nil {
			t.Fatalf("locator %d has nil page adjunct despite current pagination", i)
		}
		want := golden[pl.Locator.UnitID]
		if pl.Page.Page != want {
			t.Fatalf("locator %d page = %d, want golden %d", i, pl.Page.Page, want)
		}
	}
}
