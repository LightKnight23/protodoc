package render

import "testing"

// TestFR_098_ReflowNoScrollBelow320px is T-0251's named integration test
// (FR-098). Reflow succeeds and requires zero horizontal scrolling for every
// viewport width in [320, 1920] reference pixels across a representative
// corpus, given no declared 2-D regions (the default case).
func TestFR_098_ReflowNoScrollBelow320px(t *testing.T) {
	// Representative corpus: paragraphs whose widest word fits within the
	// minimum viewport (320 px * 1000 base units = 320000 base units).
	corpus := []BreakTable{
		buildParagraph(40, 40_000, 12_000),  // ~40 short words
		buildParagraph(120, 25_000, 8_000),  // many small words
		buildParagraph(15, 200_000, 30_000), // a few wide words (still < 320000)
	}

	for docIdx, tbl := range corpus {
		for w := 320; w <= 1920; w += 16 {
			fit := FitAtViewport(tbl, w)
			if !fit.Succeeded {
				t.Fatalf("doc %d width %d: reflow did not succeed", docIdx, w)
			}
			if fit.RequiresHScroll {
				t.Errorf("doc %d width %d: requires horizontal scrolling (max line %d > %d), no 2-D region declared",
					docIdx, w, fit.MaxLineWidth, w*pixelBaseUnits)
			}
		}
	}

	// Below the floor, reflow is out of contract (not claimed to succeed).
	if FitAtViewport(corpus[0], 319).Succeeded {
		t.Error("reflow must not claim success below the 320-pixel floor")
	}
	// Exactly at the floor, it succeeds.
	if !FitAtViewport(corpus[0], MinViewportPixels).Succeeded {
		t.Error("reflow must succeed at exactly 320 reference pixels")
	}
}
