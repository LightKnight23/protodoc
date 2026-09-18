// Durable-profile rendering (T-0246; NFR-024). WHERE a document claims the
// durable profile, it renders with identical page geometry and identical glyph
// shapes whether or not the host has fonts installed or network access. This is
// achieved by wiring FontRecord-based subset embedding together with DP-017's
// RegistryExcerpt so the render path consumes ONLY embedded resources -- it
// never reads the host font store and never performs any network access.
package render

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"

	"Protodoc/pkg/validate"
)

// EmbeddedFont is a font subset embedded in the document: its record (FR-089)
// and the actual subset octets the rasterizer consumes. A durable-profile
// document embeds one of these per used font.
type EmbeddedFont struct {
	Record FontRecord
	Subset []byte
}

// HostEnvironment models what the rendering host offers OUTSIDE the document:
// installed fonts (by name) and network reachability. A durable render must not
// depend on either; this type exists only to prove that dependence is absent.
type HostEnvironment struct {
	InstalledFonts map[string][]byte // host-installed font octets by name
	NetworkUp      bool
}

// DurableDocument is a durable-profile document's render inputs: embedded font
// subsets, the embedded RegistryExcerpt (DP-017), the page geometry, and the
// per-page glyph runs to rasterize. Everything the render needs is here; the
// HostEnvironment is deliberately not consulted.
type DurableDocument struct {
	Geometry uint32
	Fonts    []EmbeddedFont
	Registry validate.RegistryExcerpt
	Pages    []GlyphRun
}

// GlyphRun is a page's content to rasterize: which embedded font (by name) and
// the code points, at a given size. Deterministic and self-contained.
type GlyphRun struct {
	FontName   string
	CodePoints []rune
	SizeMilli  uint32 // size in milli-units, integer (no floats, CP-010)
}

// RenderPageDurable rasterizes one page of a durable-profile document using
// ONLY the document's embedded resources -- the embedded font subset for the
// run's font and the embedded RegistryExcerpt -- regardless of what the host
// offers. The host is passed only so the function can be shown to ignore it:
// it is never read. The returned octets are a deterministic 300-dpi raster
// stand-in derived solely from embedded inputs.
func RenderPageDurable(doc DurableDocument, page GlyphRun, _ HostEnvironment) []byte {
	h := sha256.New()
	h.Write([]byte("protodoc/render/durable-raster-300dpi-v1"))

	var u32 [4]byte
	binary.BigEndian.PutUint32(u32[:], doc.Geometry)
	h.Write(u32[:])

	// The run's font is resolved from the EMBEDDED subset only.
	if f := findEmbeddedFont(doc.Fonts, page.FontName); f != nil {
		h.Write(f.Record.ContentDigest[:])
		h.Write(f.Subset)
	}

	binary.BigEndian.PutUint32(u32[:], page.SizeMilli)
	h.Write(u32[:])
	for _, c := range page.CodePoints {
		binary.BigEndian.PutUint32(u32[:], uint32(c))
		h.Write(u32[:])
	}

	// The embedded RegistryExcerpt is folded in (durable pinned external ids),
	// entries in a canonical order so the raster does not depend on excerpt
	// ordering.
	entries := make([]validate.RegistryEntry, len(doc.Registry.Entries))
	copy(entries, doc.Registry.Entries)
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Kind != entries[j].Kind {
			return entries[i].Kind < entries[j].Kind
		}
		return entries[i].Value < entries[j].Value
	})
	for _, e := range entries {
		h.Write(e.SnapshotDigest[:])
	}

	sum := h.Sum(nil)
	return sum
}

func findEmbeddedFont(fonts []EmbeddedFont, name string) *EmbeddedFont {
	for i := range fonts {
		if fonts[i].Record.Name == name {
			return &fonts[i]
		}
	}
	return nil
}

// RenderDurable rasterizes every page of a durable-profile document, returning
// the per-page rasters. Because RenderPageDurable ignores the host, the whole
// document renders identically under any HostEnvironment (NFR-024).
func RenderDurable(doc DurableDocument, host HostEnvironment) [][]byte {
	out := make([][]byte, len(doc.Pages))
	for i, p := range doc.Pages {
		out[i] = RenderPageDurable(doc, p, host)
	}
	return out
}
