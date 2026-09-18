// Missing-resource placeholder (T-0258; FR-111). IF a font, image, or other
// embedded resource required for rendering is absent, fails its digest check,
// or fails to decode, a conforming reader renders in its place a mark occupying
// EXACTLY the resource's recorded extent, producing at least one non-background
// sample within that extent, classified as non-decorative and carrying a text
// alternative naming the missing resource. Preserving the recorded extent keeps
// pagination and any pinned presentation stable while making the loss visible.
package render

import "Protodoc/pkg/pdlfmt"

// backgroundSample is the raster background value (white). A placeholder must
// produce at least one sample that differs from this within its extent.
const backgroundSample = 0xFF

// Placeholder is the render of a failed/absent resource: a mark at the recorded
// extent, non-decorative, with a text alternative naming the missing resource.
type Placeholder struct {
	Resource        pdlfmt.UnitID
	Extent          Extent
	Decorative      bool   // always false: a substitution mark is non-decorative
	TextAlternative string // names the missing resource (stub pending M15 alt-text)
	// Raster is the placeholder's RGB raster at the recorded extent, row-major.
	Raster []byte
}

// RenderPlaceholder produces the FR-111 placeholder for a resource that could
// not be rendered, at exactly its recorded extent. The raster is filled with
// background except for a non-background diagonal mark guaranteeing at least one
// non-background sample; the mark is classified non-decorative and carries a
// text alternative naming the resource.
func RenderPlaceholder(resource pdlfmt.UnitID, extent Extent) Placeholder {
	w, h := extent.Width, extent.Height
	raster := make([]byte, w*h*3)
	for i := range raster {
		raster[i] = backgroundSample
	}
	// Draw a non-background diagonal (and the extent's border corners) so at
	// least one -- in fact a visible set of -- non-background samples exist.
	for d := 0; d < w && d < h; d++ {
		o := (d*w + d) * 3
		raster[o] = 0x00 // R
		raster[o+1] = 0x00
		raster[o+2] = 0x00
	}
	if w > 0 && h > 0 {
		// Guarantee a mark even for a 1x1 extent.
		raster[0], raster[1], raster[2] = 0x00, 0x00, 0x00
	}

	return Placeholder{
		Resource:        resource,
		Extent:          extent,
		Decorative:      false,
		TextAlternative: "missing resource " + hexID(resource),
		Raster:          raster,
	}
}

// HasNonBackgroundSample reports whether the placeholder raster contains at
// least one sample differing from the background (FR-111 requires >= 1).
func (p Placeholder) HasNonBackgroundSample() bool {
	for _, s := range p.Raster {
		if s != backgroundSample {
			return true
		}
	}
	return false
}

// hexID renders a unit id as hex for the text alternative.
func hexID(id pdlfmt.UnitID) string {
	const hexdig = "0123456789abcdef"
	b := make([]byte, len(id)*2)
	for i, v := range id {
		b[i*2] = hexdig[v>>4]
		b[i*2+1] = hexdig[v&0x0f]
	}
	return string(b)
}
