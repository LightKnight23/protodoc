package render

import (
	"strings"
	"testing"
)

// TestFR_111_MissingResourceRendersPlaceholder is T-0258's named integration
// test (FR-111). Under a font-less/network-less harness, a missing/failed
// resource is replaced by a mark occupying exactly the recorded extent, with at
// least one non-background sample, classified non-decorative, carrying a text
// alternative naming the missing resource.
func TestFR_111_MissingResourceRendersPlaceholder(t *testing.T) {
	extents := []Extent{
		{Width: 1, Height: 1},
		{Width: 8, Height: 4},
		{Width: 100, Height: 60},
	}
	for _, ext := range extents {
		res := pdUnit(0x5A)
		ph := RenderPlaceholder(res, ext)

		// Occupies exactly the recorded extent.
		if ph.Extent != ext {
			t.Errorf("placeholder extent %+v, want %+v", ph.Extent, ext)
		}
		if len(ph.Raster) != ext.Width*ext.Height*3 {
			t.Errorf("placeholder raster %d octets, want %d", len(ph.Raster), ext.Width*ext.Height*3)
		}
		// At least one non-background sample.
		if !ph.HasNonBackgroundSample() {
			t.Errorf("extent %+v: placeholder has no non-background sample", ext)
		}
		// Non-decorative.
		if ph.Decorative {
			t.Errorf("placeholder must be classified non-decorative")
		}
		// Text alternative names the missing resource.
		if !strings.Contains(ph.TextAlternative, hexID(res)) {
			t.Errorf("text alternative %q must name the missing resource %x", ph.TextAlternative, res)
		}
	}
}
