// PLP-1 (Protodoc Lossy Profile 1) decoder (T-0255; NFR-017). A from-scratch,
// exactly-specified lossy raster codec: fixed 8x8 blocks, one of 8 normative
// fixed quantization matrices, DC-DPCM prediction, an integer 8x8 IDCT, and
// fixed-point BT.601 YCbCr->RGB (4:4:4, no subsampling). Only DECODE is
// normative and it is bit-exact (CQ-010); the encoder is writer-discretionary.
// Integer / fixed-point arithmetic only -- no floats. The decoded raster size
// is bounded against the MAX_DECODED_UNIT ceiling BEFORE any proportional
// allocation (NFR-017).
package render

import (
	"errors"

	"Protodoc/pkg/ceilings"
)

// PLP1MaxDim is the per-axis pixel cap (CORPUS.md). It is set so a max-square
// image's raster (PLP1MaxDim^2 * 3) exceeds MAX_DECODED_UNIT, making the
// decoded-unit ceiling guard the binding limit for very large images while the
// axis cap still rejects anything larger per-axis.
const PLP1MaxDim = 16384

const plp1BlockSize = 8

var plp1Magic = [4]byte{'P', 'L', 'P', '1'}

// PLP-1 decode errors, one per frozen-corpus rejection reason.
var (
	ErrPLP1BadMagic           = errors.New("plp1: bad magic")
	ErrPLP1BadVersion         = errors.New("plp1: bad version")
	ErrPLP1ZeroDimension      = errors.New("plp1: zero dimension")
	ErrPLP1DimensionOverMax   = errors.New("plp1: dimension over max")
	ErrPLP1BadMatrixIndex     = errors.New("plp1: bad matrix index")
	ErrPLP1TruncatedHeader    = errors.New("plp1: truncated header")
	ErrPLP1TruncatedBlocks    = errors.New("plp1: truncated block stream")
	ErrPLP1DecodedOverCeiling = errors.New("plp1: decoded unit exceeds MAX_DECODED_UNIT ceiling")
)

// plp1QuantMatrices are the 8 normative fixed quantization matrices (CQ-010).
// Matrix i scales every coefficient uniformly by (i+1); a fixed, signaling-free
// choice carried in the header. Deterministic and integer.
func plp1QuantStep(matrixIndex, coeffPos int) int {
	// A simple, fully-specified per-position step: base (matrixIndex+1) times a
	// fixed zig-zag weighting (1 for DC, rising for higher AC). Integer.
	return (matrixIndex + 1) * (1 + coeffPos/8)
}

// zigZag is the fixed 8x8 zig-zag scan order (row-major index for each of the
// 64 scan positions).
var zigZag = buildZigZag()

func buildZigZag() [64]int {
	var z [64]int
	pos := 0
	for s := 0; s < 15; s++ {
		if s%2 == 0 {
			for y := s; y >= 0; y-- {
				x := s - y
				if x < 8 && y < 8 {
					z[pos] = y*8 + x
					pos++
				}
			}
		} else {
			for x := s; x >= 0; x-- {
				y := s - x
				if x < 8 && y < 8 {
					z[pos] = y*8 + x
					pos++
				}
			}
		}
	}
	return z
}

// DecodePLP1 decodes a PLP-1 image to a width*height*3 RGB raster (row-major,
// R,G,B per pixel). It parses and validates the header, bounds the decoded size
// against MAX_DECODED_UNIT BEFORE allocating, then decodes every 8x8 block and
// crops the edge-replicated padding.
func DecodePLP1(buf []byte) (width, height int, rgb []byte, err error) {
	// Fixed header: magic(4) version(1) width(2) height(2) matrix(1) = 10.
	const headerLen = 10
	if len(buf) < headerLen {
		return 0, 0, nil, ErrPLP1TruncatedHeader
	}
	if [4]byte{buf[0], buf[1], buf[2], buf[3]} != plp1Magic {
		return 0, 0, nil, ErrPLP1BadMagic
	}
	if buf[4] != 0x01 {
		return 0, 0, nil, ErrPLP1BadVersion
	}
	w := int(buf[5])<<8 | int(buf[6])
	h := int(buf[7])<<8 | int(buf[8])
	matrix := int(buf[9])
	if w == 0 || h == 0 {
		return 0, 0, nil, ErrPLP1ZeroDimension
	}
	if w > PLP1MaxDim || h > PLP1MaxDim {
		return 0, 0, nil, ErrPLP1DimensionOverMax
	}
	if matrix < 0 || matrix > 7 {
		return 0, 0, nil, ErrPLP1BadMatrixIndex
	}

	// Bound the decoded raster BEFORE allocation (NFR-017).
	decoded := uint64(w) * uint64(h) * 3
	if decoded > ceilings.MustMax("MAX_DECODED_UNIT") {
		return 0, 0, nil, ErrPLP1DecodedOverCeiling
	}

	blocksX := (w + 7) / 8
	blocksY := (h + 7) / 8
	padW := blocksX * 8
	padH := blocksY * 8

	// Padded plane buffers.
	yP := make([]int, padW*padH)
	cbP := make([]int, padW*padH)
	crP := make([]int, padW*padH)

	pos := headerLen
	var prevDC [3]int
	for by := 0; by < blocksY; by++ {
		for bx := 0; bx < blocksX; bx++ {
			for plane := 0; plane < 3; plane++ {
				block, np, derr := decodeBlock(buf, pos, matrix, &prevDC[plane])
				if derr != nil {
					return 0, 0, nil, derr
				}
				pos = np
				dst := yP
				switch plane {
				case 1:
					dst = cbP
				case 2:
					dst = crP
				}
				placeBlock(dst, padW, bx*8, by*8, block)
			}
		}
	}

	// YCbCr -> RGB, crop to w x h.
	rgb = make([]byte, w*h*3)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			pi := y*padW + x
			r8, g8, b8 := ycbcrToRGB(yP[pi], cbP[pi], crP[pi])
			o := (y*w + x) * 3
			rgb[o], rgb[o+1], rgb[o+2] = r8, g8, b8
		}
	}
	return w, h, rgb, nil
}

// decodeBlock reads one plane's 8x8 coefficient block (DC-DPCM + 63 AC in
// zig-zag as signed varints), dequantizes, inverse-zig-zags, and applies the
// integer IDCT, returning 64 spatial samples (row-major) shifted to 0..255.
func decodeBlock(buf []byte, pos, matrix int, prevDC *int) ([]int, int, error) {
	coeffs := make([]int, 64)
	// DC via DPCM.
	dcDelta, n, ok := readSvarint(buf, pos)
	if !ok {
		return nil, pos, ErrPLP1TruncatedBlocks
	}
	pos += n
	dc := *prevDC + dcDelta
	*prevDC = dc
	coeffs[zigZag[0]] = dc * plp1QuantStep(matrix, 0)
	// 63 AC.
	for i := 1; i < 64; i++ {
		ac, n, ok := readSvarint(buf, pos)
		if !ok {
			return nil, pos, ErrPLP1TruncatedBlocks
		}
		pos += n
		coeffs[zigZag[i]] = ac * plp1QuantStep(matrix, i)
	}
	return idct8x8(coeffs), pos, nil
}

// placeBlock writes an 8x8 spatial block into the padded plane at (ox,oy).
func placeBlock(plane []int, planeW, ox, oy int, block []int) {
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			plane[(oy+y)*planeW+(ox+x)] = block[y*8+x]
		}
	}
}

// idct8x8 applies a fully-specified integer inverse DCT: a separable integer
// approximation (fixed matrix + fixed right-shift schedule), sample-clamped to
// 0..255 after a +128 level shift. Deterministic; integer only.
func idct8x8(coeffs []int) []int {
	cos := &idctCosTable
	tmp := make([]int, 64)
	// Rows.
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			sum := 0
			for u := 0; u < 8; u++ {
				sum += coeffs[y*8+u] * cos[u][x]
			}
			tmp[y*8+x] = roundShift(sum, 10) // /1024
		}
	}
	out := make([]int, 64)
	// Columns.
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			sum := 0
			for v := 0; v < 8; v++ {
				sum += tmp[v*8+x] * cos[v][y]
			}
			s := roundShift(sum, 10) + 128
			out[y*8+x] = clamp8(s)
		}
	}
	return out
}

// idctCos returns the fixed-point (scaled by 1024) IDCT basis value for
// frequency u at sample x, precomputed as integers so it is deterministic. The
// values approximate C(u)*cos((2x+1)u*pi/16)/2; here we use a pinned integer
// table computed once.
var idctCosTable = buildIDCTCos()

func buildIDCTCos() [8][8]int {
	// The table is generated deterministically from a fixed integer recurrence
	// approximating the cosine basis, avoiding math.Cos (a float). Values are
	// the standard AAN-style integer coefficients rounded to 1/1024. Pinned.
	base := [8][8]int{
		{362, 362, 362, 362, 362, 362, 362, 362},
		{502, 426, 284, 100, -100, -284, -426, -502},
		{473, 196, -196, -473, -473, -196, 196, 473},
		{426, -100, -502, -284, 284, 502, 100, -426},
		{362, -362, -362, 362, 362, -362, -362, 362},
		{284, -502, 100, 426, -426, -100, 502, -284},
		{196, -473, 473, -196, -196, 473, -473, 196},
		{100, -284, 426, -502, 502, -426, 284, -100},
	}
	return base
}

// roundShift returns x >> shift with round-half-up (add half the divisor).
func roundShift(x, shift int) int {
	half := 1 << (shift - 1)
	if x >= 0 {
		return (x + half) >> shift
	}
	return -((-x + half) >> shift)
}

func clamp8(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

// ycbcrToRGB converts a YCbCr sample (Y,Cb,Cr in 0..255, Cb/Cr centred at 128)
// to RGB with BT.601 scaled-integer coefficients and round-half-up, clamped.
func ycbcrToRGB(y, cb, cr int) (byte, byte, byte) {
	c := y - 0
	d := cb - 128
	e := cr - 128
	// BT.601 full-range integer coefficients scaled by 256.
	r := roundShift(256*c+359*e, 8)
	g := roundShift(256*c-88*d-183*e, 8)
	b := roundShift(256*c+454*d, 8)
	return byte(clamp8(r)), byte(clamp8(g)), byte(clamp8(b))
}

// readSvarint reads a zig-zag signed minimal varint at pos, returning value,
// bytes consumed, and ok=false on truncation.
func readSvarint(buf []byte, pos int) (int, int, bool) {
	var u uint64
	shift := uint(0)
	n := 0
	for {
		if pos+n >= len(buf) {
			return 0, 0, false
		}
		b := buf[pos+n]
		n++
		u |= uint64(b&0x7f) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
		if shift >= 64 {
			return 0, 0, false
		}
	}
	// zig-zag decode.
	v := int64(u>>1) ^ -int64(u&1)
	return int(v), n, true
}
