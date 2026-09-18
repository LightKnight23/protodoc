package render

import (
	"bytes"
	"testing"

	"Protodoc/pkg/pdlfmt"
	"Protodoc/pkg/validate"
)

// TestNFR_024_DurableProfileRendersIdenticallyOffline is T-0246's named
// integration test (NFR-024). A durable-profile document -- embedded font
// subsets wired with the DP-017 RegistryExcerpt -- renders to byte-identical
// rasters for every page whether the host has fonts/network available or has
// neither, because the render path consumes only embedded resources.
func TestNFR_024_DurableProfileRendersIdenticallyOffline(t *testing.T) {
	doc := DurableDocument{
		Geometry: 0x00300300, // 300-dpi geometry token
		Fonts: []EmbeddedFont{
			{
				Record: FontRecord{Name: "Durable Serif", Version: "1.0", ContentDigest: pdlfmt.Digest256{0x11}},
				Subset: []byte("embedded-serif-subset-octets"),
			},
			{
				Record: FontRecord{Name: "Durable Mono", Version: "2.0", ContentDigest: pdlfmt.Digest256{0x22}},
				Subset: []byte("embedded-mono-subset-octets"),
			},
		},
		Registry: validate.RegistryExcerpt{Entries: []validate.RegistryEntry{
			{Value: 7, SnapshotDigest: [32]byte{0xAA}},
			{Value: 3, SnapshotDigest: [32]byte{0xBB}},
		}},
		Pages: []GlyphRun{
			{FontName: "Durable Serif", CodePoints: []rune("Hello"), SizeMilli: 12000},
			{FontName: "Durable Mono", CodePoints: []rune("World"), SizeMilli: 10000},
		},
	}

	// Harness A: host fonts installed AND network up.
	hostAvailable := HostEnvironment{
		InstalledFonts: map[string][]byte{
			"Durable Serif": []byte("DIFFERENT host serif octets"),
			"Durable Mono":  []byte("DIFFERENT host mono octets"),
		},
		NetworkUp: true,
	}
	// Harness B: no fonts installed, no network.
	hostAbsent := HostEnvironment{InstalledFonts: nil, NetworkUp: false}

	rastersA := RenderDurable(doc, hostAvailable)
	rastersB := RenderDurable(doc, hostAbsent)

	if len(rastersA) != len(doc.Pages) || len(rastersB) != len(doc.Pages) {
		t.Fatalf("rendered %d/%d pages, want %d", len(rastersA), len(rastersB), len(doc.Pages))
	}

	// 100% page equality between the two harnesses (NFR-024).
	for i := range rastersA {
		if !bytes.Equal(rastersA[i], rastersB[i]) {
			t.Errorf("page %d raster differs between host-available and host-absent harnesses", i)
		}
	}

	// Sanity: the raster actually depends on the EMBEDDED subset (so equality
	// is meaningful, not because the raster ignores fonts entirely). Changing
	// an embedded subset changes the raster.
	doc2 := doc
	doc2.Fonts = append([]EmbeddedFont(nil), doc.Fonts...)
	doc2.Fonts[0].Subset = []byte("a different embedded subset")
	if bytes.Equal(RenderDurable(doc2, hostAbsent)[0], rastersB[0]) {
		t.Error("raster must depend on the embedded font subset")
	}
}
