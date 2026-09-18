package render

import (
	"errors"
	"testing"
)

// TestFR_090_UnimplementableObjectUsesMaterialisedRepresentation is T-0257's
// named integration test (FR-090). A reader implementing NONE of the native
// object kinds renders every embedded object at its declared extent from the
// stored materialised representation (PLP-1 or restricted-PNG), and a
// representation whose decoded extent disagrees with the declared extent is
// rejected.
func TestFR_090_UnimplementableObjectUsesMaterialisedRepresentation(t *testing.T) {
	// A corpus of "unknown" native kinds, each with a materialised raster.
	plp1Payload := buildValidVector("gradient_8x8", 8, 8, 3) // 8x8 materialised
	pngPayload := encodeRestrictedPNG(2, 2, 3, []byte{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
	})

	objs := []EmbeddedObject{
		{ID: pdUnit(0x01), NativeKind: 0xDEAD0001, DeclExtent: Extent{Width: 8, Height: 8}, Materialised: MaterialisedPLP1, Payload: plp1Payload},
		{ID: pdUnit(0x02), NativeKind: 0xDEAD0002, DeclExtent: Extent{Width: 2, Height: 2}, Materialised: MaterialisedRestrictedPNG, Payload: pngPayload},
	}

	for _, obj := range objs {
		got, err := RenderMaterialised(obj)
		if err != nil {
			t.Errorf("obj %x (native kind %#x): materialised render failed: %v", obj.ID, obj.NativeKind, err)
			continue
		}
		// Rendered at exactly the declared extent, independent of native kind.
		if got.Extent != obj.DeclExtent {
			t.Errorf("obj %x: rendered extent %+v, want declared %+v", obj.ID, got.Extent, obj.DeclExtent)
		}
		if len(got.Raster) != obj.DeclExtent.Width*obj.DeclExtent.Height*got.Channels {
			t.Errorf("obj %x: raster %d octets, want %d", obj.ID, len(got.Raster), obj.DeclExtent.Width*obj.DeclExtent.Height*got.Channels)
		}
	}

	// A materialised representation whose decoded extent disagrees with the
	// declared extent is rejected (FR-090 requires equal extents).
	mismatch := EmbeddedObject{
		ID: pdUnit(0x03), NativeKind: 1,
		DeclExtent:   Extent{Width: 16, Height: 16}, // declared 16x16
		Materialised: MaterialisedPLP1,
		Payload:      buildValidVector("gradient_8x8", 8, 8, 3), // but raster is 8x8
	}
	if _, err := RenderMaterialised(mismatch); !errors.Is(err, ErrMaterialisedExtentMismatch) {
		t.Errorf("extent mismatch: err = %v, want ErrMaterialisedExtentMismatch", err)
	}
}
