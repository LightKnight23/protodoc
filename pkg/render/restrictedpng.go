// Restricted-PNG lossless decoder (T-0256; NFR-017). One narrow, exactly-
// specified lossless raster profile: it accepts ONLY the allowed chunk and
// colour-type subset and rejects everything else, decoding to a pixel-identical
// raster. The allowed subset is:
//   - signature: the 8-octet PNG signature.
//   - chunks: exactly IHDR, one or more IDAT (contiguous), IEND. No ancillary
//     chunks (tEXt, gAMA, etc.), no PLTE, no interlacing.
//   - colour type: 2 (truecolour RGB) or 6 (truecolour+alpha), bit depth 8.
//   - filter method 0, compression method 0, interlace method 0.
//
// IDAT is inflated with the stdlib zlib reader; scanline un-filtering is done
// here per the PNG filter algorithms. The decoded size is bounded against
// MAX_DECODED_UNIT before allocation.
package render

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"io"

	"Protodoc/pkg/ceilings"
)

var pngSignature = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}

// Restricted-PNG decode errors.
var (
	ErrPNGBadSignature    = errors.New("restrictedpng: bad signature")
	ErrPNGTruncated       = errors.New("restrictedpng: truncated")
	ErrPNGDisallowedChunk = errors.New("restrictedpng: disallowed chunk")
	ErrPNGBadColorType    = errors.New("restrictedpng: colour type not in {2,6}")
	ErrPNGBadBitDepth     = errors.New("restrictedpng: bit depth must be 8")
	ErrPNGInterlace       = errors.New("restrictedpng: interlacing not allowed")
	ErrPNGBadMethod       = errors.New("restrictedpng: filter/compression method not 0")
	ErrPNGChunkOrder      = errors.New("restrictedpng: chunk order invalid")
	ErrPNGOverCeiling     = errors.New("restrictedpng: decoded unit exceeds MAX_DECODED_UNIT ceiling")
	ErrPNGBadFilter       = errors.New("restrictedpng: unknown scanline filter type")
	ErrPNGZlib            = errors.New("restrictedpng: zlib decompression failed")
)

// RestrictedPNGImage is the decoded result: dimensions, channel count (3 or 4),
// and the un-filtered raster (row-major, channels per pixel, 8-bit).
type RestrictedPNGImage struct {
	Width, Height int
	Channels      int
	Pixels        []byte
}

// DecodeRestrictedPNG decodes a restricted-PNG buffer, rejecting any chunk,
// colour type, bit depth, or method outside the allowed subset.
func DecodeRestrictedPNG(buf []byte) (*RestrictedPNGImage, error) {
	if len(buf) < len(pngSignature) || !bytes.Equal(buf[:8], pngSignature) {
		return nil, ErrPNGBadSignature
	}
	pos := 8

	var (
		width, height int
		channels      int
		sawIHDR       bool
		sawIEND       bool
		idatDone      bool
		idat          []byte
	)

	for pos < len(buf) {
		if pos+8 > len(buf) {
			return nil, ErrPNGTruncated
		}
		length := int(binary.BigEndian.Uint32(buf[pos : pos+4]))
		ctype := string(buf[pos+4 : pos+8])
		pos += 8
		if length < 0 || pos+length+4 > len(buf) {
			return nil, ErrPNGTruncated
		}
		data := buf[pos : pos+length]
		pos += length + 4 // skip CRC (not verified in this profile subset)

		switch ctype {
		case "IHDR":
			if sawIHDR {
				return nil, ErrPNGChunkOrder
			}
			sawIHDR = true
			if length != 13 {
				return nil, ErrPNGTruncated
			}
			width = int(binary.BigEndian.Uint32(data[0:4]))
			height = int(binary.BigEndian.Uint32(data[4:8]))
			bitDepth := data[8]
			colorType := data[9]
			compMethod := data[10]
			filterMethod := data[11]
			interlace := data[12]
			if width == 0 || height == 0 {
				return nil, ErrPNGTruncated
			}
			if bitDepth != 8 {
				return nil, ErrPNGBadBitDepth
			}
			switch colorType {
			case 2:
				channels = 3
			case 6:
				channels = 4
			default:
				return nil, ErrPNGBadColorType
			}
			if compMethod != 0 || filterMethod != 0 {
				return nil, ErrPNGBadMethod
			}
			if interlace != 0 {
				return nil, ErrPNGInterlace
			}
			// Bound decoded size before any large allocation.
			decoded := uint64(width) * uint64(height) * uint64(channels)
			if decoded > ceilings.MustMax("MAX_DECODED_UNIT") {
				return nil, ErrPNGOverCeiling
			}
		case "IDAT":
			if !sawIHDR || idatDone {
				return nil, ErrPNGChunkOrder
			}
			idat = append(idat, data...)
		case "IEND":
			if !sawIHDR {
				return nil, ErrPNGChunkOrder
			}
			sawIEND = true
			pos = len(buf) // IEND terminates
		default:
			// Any chunk outside the allowed subset (PLTE, tEXt, gAMA, ...) is
			// rejected -- restricted profile.
			return nil, ErrPNGDisallowedChunk
		}
		if ctype == "IDAT" {
			// mark that subsequent non-IDAT ends the IDAT run.
		} else if ctype != "IHDR" && ctype != "IEND" {
			idatDone = true
		}
	}
	if !sawIHDR || !sawIEND {
		return nil, ErrPNGChunkOrder
	}

	raw, err := inflate(idat)
	if err != nil {
		return nil, err
	}
	pixels, err := unfilter(raw, width, height, channels)
	if err != nil {
		return nil, err
	}
	return &RestrictedPNGImage{Width: width, Height: height, Channels: channels, Pixels: pixels}, nil
}

func inflate(idat []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(idat))
	if err != nil {
		return nil, ErrPNGZlib
	}
	defer r.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		return nil, ErrPNGZlib
	}
	return out, nil
}

// unfilter reverses the PNG scanline filters (None/Sub/Up/Average/Paeth),
// producing the row-major raster. Each scanline is prefixed by a filter-type
// octet.
func unfilter(raw []byte, width, height, channels int) ([]byte, error) {
	stride := width * channels
	if len(raw) != height*(stride+1) {
		return nil, ErrPNGTruncated
	}
	out := make([]byte, height*stride)
	prev := make([]byte, stride)
	for y := 0; y < height; y++ {
		ft := raw[y*(stride+1)]
		row := raw[y*(stride+1)+1 : y*(stride+1)+1+stride]
		cur := out[y*stride : y*stride+stride]
		copy(cur, row)
		switch ft {
		case 0: // None
		case 1: // Sub
			for i := 0; i < stride; i++ {
				var a byte
				if i >= channels {
					a = cur[i-channels]
				}
				cur[i] += a
			}
		case 2: // Up
			for i := 0; i < stride; i++ {
				cur[i] += prev[i]
			}
		case 3: // Average
			for i := 0; i < stride; i++ {
				var a byte
				if i >= channels {
					a = cur[i-channels]
				}
				cur[i] += byte((int(a) + int(prev[i])) / 2)
			}
		case 4: // Paeth
			for i := 0; i < stride; i++ {
				var a, c byte
				if i >= channels {
					a = cur[i-channels]
					c = prev[i-channels]
				}
				b := prev[i]
				cur[i] += paeth(a, b, c)
			}
		default:
			return nil, ErrPNGBadFilter
		}
		copy(prev, cur)
	}
	return out, nil
}

func paeth(a, b, c byte) byte {
	p := int(a) + int(b) - int(c)
	pa := abs(p - int(a))
	pb := abs(p - int(b))
	pc := abs(p - int(c))
	switch {
	case pa <= pb && pa <= pc:
		return a
	case pb <= pc:
		return b
	default:
		return c
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
