package render

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"testing"
)

// writeChunk appends a PNG chunk (length, type, data, CRC) to buf.
func writeChunk(buf *bytes.Buffer, ctype string, data []byte) {
	var lb [4]byte
	binary.BigEndian.PutUint32(lb[:], uint32(len(data)))
	buf.Write(lb[:])
	buf.WriteString(ctype)
	buf.Write(data)
	crc := crc32.NewIEEE()
	crc.Write([]byte(ctype))
	crc.Write(data)
	var cb [4]byte
	binary.BigEndian.PutUint32(cb[:], crc.Sum32())
	buf.Write(cb[:])
}

// encodeRestrictedPNG builds a restricted PNG (colour type 2 or 6, bit depth 8,
// filter type 0 per scanline) from a raster. Discretionary encoder; only decode
// is normative.
func encodeRestrictedPNG(width, height, channels int, pixels []byte) []byte {
	var buf bytes.Buffer
	buf.Write(pngSignature)

	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr[0:4], uint32(width))
	binary.BigEndian.PutUint32(ihdr[4:8], uint32(height))
	ihdr[8] = 8 // bit depth
	if channels == 4 {
		ihdr[9] = 6
	} else {
		ihdr[9] = 2
	}
	// comp=0 filter=0 interlace=0
	writeChunk(&buf, "IHDR", ihdr)

	stride := width * channels
	var raw bytes.Buffer
	for y := 0; y < height; y++ {
		raw.WriteByte(0) // filter type None
		raw.Write(pixels[y*stride : y*stride+stride])
	}
	var zbuf bytes.Buffer
	zw := zlib.NewWriter(&zbuf)
	zw.Write(raw.Bytes())
	zw.Close()
	writeChunk(&buf, "IDAT", zbuf.Bytes())
	writeChunk(&buf, "IEND", nil)
	return buf.Bytes()
}

// TestRestrictedPNG_DecodeConformance is T-0256's named conformance test
// (NFR-017). The decoder round-trips a golden restricted-PNG corpus bit-exact
// and rejects every disallowed chunk/colour-type/method as a conformance
// fixture.
func TestRestrictedPNG_DecodeConformance(t *testing.T) {
	// Golden corpus: RGB and RGBA rasters, decode must be pixel-identical.
	golden := []struct {
		name           string
		w, h, channels int
		pixels         []byte
	}{
		{"rgb_2x2", 2, 2, 3, []byte{
			0xFF, 0x00, 0x00, 0x00, 0xFF, 0x00,
			0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF,
		}},
		{"rgba_1x3", 1, 3, 4, []byte{
			0x10, 0x20, 0x30, 0x40,
			0x50, 0x60, 0x70, 0x80,
			0x90, 0xA0, 0xB0, 0xC0,
		}},
		{"rgb_3x1_gradient", 3, 1, 3, []byte{
			0x00, 0x00, 0x00, 0x7F, 0x7F, 0x7F, 0xFF, 0xFF, 0xFF,
		}},
	}
	for _, g := range golden {
		enc := encodeRestrictedPNG(g.w, g.h, g.channels, g.pixels)
		img, err := DecodeRestrictedPNG(enc)
		if err != nil {
			t.Errorf("%s: valid restricted PNG rejected: %v", g.name, err)
			continue
		}
		if img.Width != g.w || img.Height != g.h || img.Channels != g.channels {
			t.Errorf("%s: decoded %dx%dx%d, want %dx%dx%d", g.name, img.Width, img.Height, img.Channels, g.w, g.h, g.channels)
		}
		if !bytes.Equal(img.Pixels, g.pixels) {
			t.Errorf("%s: raster not pixel-identical\n got %v\nwant %v", g.name, img.Pixels, g.pixels)
		}
	}

	// Rejection fixtures.
	base := encodeRestrictedPNG(2, 2, 3, make([]byte, 12))

	// Bad signature.
	bad := append([]byte(nil), base...)
	bad[1] = 'X'
	if _, err := DecodeRestrictedPNG(bad); !errors.Is(err, ErrPNGBadSignature) {
		t.Errorf("bad signature: err = %v, want ErrPNGBadSignature", err)
	}

	// Disallowed chunk (tEXt inserted after IHDR).
	if _, err := DecodeRestrictedPNG(withExtraChunk(2, 2, 3, "tEXt", []byte("kw\x00v"))); !errors.Is(err, ErrPNGDisallowedChunk) {
		t.Errorf("tEXt chunk: err = %v, want ErrPNGDisallowedChunk", err)
	}
	// PLTE is also disallowed in the truecolour-only profile.
	if _, err := DecodeRestrictedPNG(withExtraChunk(2, 2, 3, "PLTE", []byte{1, 2, 3})); !errors.Is(err, ErrPNGDisallowedChunk) {
		t.Errorf("PLTE chunk: err = %v, want ErrPNGDisallowedChunk", err)
	}

	// Bad colour type (0 = greyscale, not allowed).
	if _, err := DecodeRestrictedPNG(withColorType(2, 2, 0)); !errors.Is(err, ErrPNGBadColorType) {
		t.Errorf("colour type 0: err = %v, want ErrPNGBadColorType", err)
	}

	// Bad bit depth (16).
	if _, err := DecodeRestrictedPNG(withBitDepth(2, 2, 16)); !errors.Is(err, ErrPNGBadBitDepth) {
		t.Errorf("bit depth 16: err = %v, want ErrPNGBadBitDepth", err)
	}

	// Interlaced.
	if _, err := DecodeRestrictedPNG(withInterlace(2, 2)); !errors.Is(err, ErrPNGInterlace) {
		t.Errorf("interlace: err = %v, want ErrPNGInterlace", err)
	}
}

// withExtraChunk builds a valid restricted PNG with an extra disallowed chunk
// inserted right after IHDR.
func withExtraChunk(w, h, channels int, ctype string, data []byte) []byte {
	valid := encodeRestrictedPNG(w, h, channels, make([]byte, w*h*channels))
	// IHDR ends at signature(8)+chunk(8+13+4)=33. Insert the extra chunk there.
	const ihdrEnd = 8 + 8 + 13 + 4
	var mid bytes.Buffer
	writeChunk(&mid, ctype, data)
	out := append([]byte(nil), valid[:ihdrEnd]...)
	out = append(out, mid.Bytes()...)
	out = append(out, valid[ihdrEnd:]...)
	return out
}

func withColorType(w, h int, ct byte) []byte {
	v := encodeRestrictedPNG(w, h, 3, make([]byte, w*h*3))
	v[8+8+9] = ct // IHDR data byte 9 = colour type
	return v
}

func withBitDepth(w, h int, bd byte) []byte {
	v := encodeRestrictedPNG(w, h, 3, make([]byte, w*h*3))
	v[8+8+8] = bd // IHDR data byte 8 = bit depth
	return v
}

func withInterlace(w, h int) []byte {
	v := encodeRestrictedPNG(w, h, 3, make([]byte, w*h*3))
	v[8+8+12] = 1 // IHDR data byte 12 = interlace method
	return v
}
