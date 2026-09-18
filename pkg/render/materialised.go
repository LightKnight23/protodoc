// Materialised representation for unimplementable objects (T-0257; FR-090).
// Every embedded object whose kind a reader may not implement carries a
// MATERIALISED rendered representation (PLP-1 or restricted-PNG) whose extent
// equals the object's declared extent. A reader that implements NONE of the
// native object kinds still renders every object at its declared extent from
// the materialised raster, independent of the native format's support.
package render

import (
	"errors"

	"Protodoc/pkg/pdlfmt"
)

// Extent is an object's declared authored placement (integer base units).
type Extent struct {
	X, Y, Width, Height int
}

// MaterialisedKind is the codec of a materialised representation.
type MaterialisedKind uint8

const (
	// MaterialisedPLP1: a PLP-1 lossy raster.
	MaterialisedPLP1 MaterialisedKind = iota
	// MaterialisedRestrictedPNG: a restricted-PNG lossless raster.
	MaterialisedRestrictedPNG
)

// EmbeddedObject is an embedded object of a possibly-unimplementable kind. It
// always carries a materialised rendered representation at its declared extent.
type EmbeddedObject struct {
	ID           pdlfmt.UnitID
	NativeKind   uint32 // the native object kind (may be unknown to the reader)
	DeclExtent   Extent
	Materialised MaterialisedKind
	Payload      []byte // the PLP-1 or restricted-PNG octets
}

// ErrMaterialisedExtentMismatch is returned when a materialised representation's
// decoded raster does not match the object's declared extent (FR-090).
var ErrMaterialisedExtentMismatch = errors.New("render: materialised representation extent != declared extent")

// RenderedObject is the result of rendering an embedded object from its
// materialised representation: its declared extent and the decoded raster (RGB
// or RGBA octets, row-major).
type RenderedObject struct {
	Extent   Extent
	Channels int
	Raster   []byte
}

// RenderMaterialised renders an embedded object from its materialised
// representation WITHOUT any native decoder: it decodes the PLP-1 or
// restricted-PNG payload and requires the decoded raster's dimensions to equal
// the declared extent (FR-090). The native kind is never consulted, so a reader
// implementing none of the native kinds renders every object identically.
func RenderMaterialised(obj EmbeddedObject) (RenderedObject, error) {
	switch obj.Materialised {
	case MaterialisedPLP1:
		w, h, rgb, err := DecodePLP1(obj.Payload)
		if err != nil {
			return RenderedObject{}, err
		}
		if w != obj.DeclExtent.Width || h != obj.DeclExtent.Height {
			return RenderedObject{}, ErrMaterialisedExtentMismatch
		}
		return RenderedObject{Extent: obj.DeclExtent, Channels: 3, Raster: rgb}, nil
	case MaterialisedRestrictedPNG:
		img, err := DecodeRestrictedPNG(obj.Payload)
		if err != nil {
			return RenderedObject{}, err
		}
		if img.Width != obj.DeclExtent.Width || img.Height != obj.DeclExtent.Height {
			return RenderedObject{}, ErrMaterialisedExtentMismatch
		}
		return RenderedObject{Extent: obj.DeclExtent, Channels: img.Channels, Raster: img.Pixels}, nil
	default:
		return RenderedObject{}, errors.New("render: unknown materialised kind")
	}
}
