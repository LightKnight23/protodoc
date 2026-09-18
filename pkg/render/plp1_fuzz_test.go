package render

import "testing"

// FuzzPLP1Decode is T-0265's named fuzz target (NFR-017 / CP-012). It feeds
// untrusted bytes to BOTH untrusted-input decoders that implement the
// page-render memory budget -- the PLP-1 decoder (T-0255) and the
// restricted-PNG decoder (T-0256) -- and requires neither to panic, OOM, or
// allocate unboundedly on any input. Both are registered here so M19's
// harness-maturity check discovers the target. The ceiling guards in both
// decoders keep allocation bounded regardless of the declared dimensions.
func FuzzPLP1Decode(f *testing.F) {
	// >= 2 seeds (fuzzmaturity requires it): a valid PLP-1 image and a valid
	// restricted-PNG image, plus a few malformed shapes.
	f.Add(buildValidVector("gradient_8x8", 8, 8, 3))
	f.Add(encodeRestrictedPNG(2, 2, 3, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}))
	f.Add([]byte("PLP1"))              // truncated header
	f.Add([]byte{0x89, 'P', 'N', 'G'}) // truncated PNG signature
	f.Add([]byte{})                    // empty

	f.Fuzz(func(t *testing.T, data []byte) {
		// Neither decoder may panic or allocate unboundedly on untrusted input.
		// The ceiling guards bound allocation; an error return is fine.
		if w, h, rgb, err := DecodePLP1(data); err == nil {
			// A successful decode must be self-consistent.
			if len(rgb) != w*h*3 {
				t.Fatalf("PLP1 decode inconsistent: %dx%d but %d octets", w, h, len(rgb))
			}
		}
		if img, err := DecodeRestrictedPNG(data); err == nil {
			if len(img.Pixels) != img.Width*img.Height*img.Channels {
				t.Fatalf("PNG decode inconsistent: %dx%dx%d but %d octets", img.Width, img.Height, img.Channels, len(img.Pixels))
			}
		}
	})
}
