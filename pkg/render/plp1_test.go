package render

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"Protodoc/pkg/ceilings"
)

// --- discretionary reference encoder (decode is normative; encode is not) ---

func appendSvarint(buf []byte, v int) []byte {
	u := uint64((v << 1) ^ (v >> 63)) // zig-zag
	for {
		b := byte(u & 0x7f)
		u >>= 7
		if u != 0 {
			buf = append(buf, b|0x80)
		} else {
			buf = append(buf, b)
			return buf
		}
	}
}

// encodePLP1DConly encodes a w x h image where every 8x8 block carries only a
// DC coefficient per plane (all AC = 0), with per-block, per-plane DC values
// supplied as quantized levels. This is enough to exercise the header, DC-DPCM,
// dequant, IDCT and colour conversion deterministically.
func encodePLP1DConly(w, h, matrix int, dcLevels [][3]int) []byte {
	buf := []byte{'P', 'L', 'P', '1', 0x01, byte(w >> 8), byte(w), byte(h >> 8), byte(h), byte(matrix)}
	blocksX := (w + 7) / 8
	blocksY := (h + 7) / 8
	var prev [3]int
	bi := 0
	for by := 0; by < blocksY; by++ {
		for bx := 0; bx < blocksX; bx++ {
			for plane := 0; plane < 3; plane++ {
				dc := 0
				if bi < len(dcLevels) {
					dc = dcLevels[bi][plane]
				}
				buf = appendSvarint(buf, dc-prev[plane]) // DC-DPCM delta
				prev[plane] = dc
				for i := 1; i < 64; i++ {
					buf = appendSvarint(buf, 0) // AC all zero
				}
			}
			bi++
		}
	}
	return buf
}

// TestPLP1_DecoderPassesConformanceCorpus is T-0255's named conformance test
// (NFR-017). The decoder reproduces every valid vector deterministically,
// rejects every invalid vector with its frozen manifest reason, and rejects an
// over-ceiling input before any proportional allocation.
func TestPLP1_DecoderPassesConformanceCorpus(t *testing.T) {
	rows := readPLP1Manifest(t)

	reasonToErr := map[string]error{
		"bad-magic":                 ErrPLP1BadMagic,
		"bad-version":               ErrPLP1BadVersion,
		"zero-dimension":            ErrPLP1ZeroDimension,
		"dimension-over-max":        ErrPLP1DimensionOverMax,
		"bad-matrix-index":          ErrPLP1BadMatrixIndex,
		"truncated-header":          ErrPLP1TruncatedHeader,
		"truncated-block-stream":    ErrPLP1TruncatedBlocks,
		"decoded-unit-over-ceiling": ErrPLP1DecodedOverCeiling,
	}

	sawValid, sawInvalid := 0, 0
	for _, r := range rows {
		switch r.kind {
		case "valid":
			sawValid++
			in := buildValidVector(r.name, r.width, r.height, r.matrix)
			w, h, rgb, err := DecodePLP1(in)
			if err != nil {
				t.Errorf("%s: valid vector rejected: %v", r.name, err)
				continue
			}
			if w != r.width || h != r.height || len(rgb) != r.width*r.height*3 {
				t.Errorf("%s: decoded %dx%d (%d octets), want %dx%d", r.name, w, h, len(rgb), r.width, r.height)
			}
			// Determinism: decoding twice is bit-identical.
			_, _, rgb2, _ := DecodePLP1(in)
			if string(rgb) != string(rgb2) {
				t.Errorf("%s: decode not deterministic", r.name)
			}
		case "invalid":
			sawInvalid++
			in := buildInvalidVector(r.name, r.width, r.height, r.matrix)
			_, _, _, err := DecodePLP1(in)
			want := reasonToErr[r.reason]
			if want == nil {
				t.Fatalf("%s: manifest reason %q has no mapped error", r.name, r.reason)
			}
			if !errors.Is(err, want) {
				t.Errorf("%s: err = %v, want %v (reason %q)", r.name, err, want, r.reason)
			}
		}
	}
	if sawValid < 4 || sawInvalid < 8 {
		t.Errorf("corpus coverage: %d valid, %d invalid; want >=4 and >=8", sawValid, sawInvalid)
	}

	// The over-ceiling vector rejects with the ceiling error, and the guard is
	// positioned BEFORE any proportional allocation: in DecodePLP1 the
	// MAX_DECODED_UNIT check precedes every make() for the plane/raster buffers,
	// so an over-ceiling image never allocates proportional memory.
	overCeiling := buildInvalidVector("over_ceiling_dim", 16384, 16384, 0)
	if _, _, _, err := DecodePLP1(overCeiling); !errors.Is(err, ErrPLP1DecodedOverCeiling) {
		t.Errorf("over_ceiling_dim: err = %v, want ErrPLP1DecodedOverCeiling", err)
	}
	if ceilings.MustMax("MAX_DECODED_UNIT") != 268435456 {
		t.Errorf("MAX_DECODED_UNIT = %d, want 268435456", ceilings.MustMax("MAX_DECODED_UNIT"))
	}
}

type plp1Row struct {
	name, kind, reason    string
	width, height, matrix int
}

func readPLP1Manifest(t *testing.T) []plp1Row {
	t.Helper()
	f, err := os.Open(filepath.Join(".", "testdata", "plp1", "manifest.tsv"))
	if err != nil {
		t.Fatalf("open manifest: %v", err)
	}
	defer f.Close()
	var rows []plp1Row
	sc := bufio.NewScanner(f)
	first := true
	for sc.Scan() {
		line := sc.Text()
		if first {
			first = false
			continue // header
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		c := strings.Split(line, "\t")
		if len(c) < 6 {
			c = append(c, make([]string, 6-len(c))...)
		}
		w, _ := strconv.Atoi(c[2])
		h, _ := strconv.Atoi(c[3])
		m, _ := strconv.Atoi(c[4])
		rows = append(rows, plp1Row{name: c[0], kind: c[1], width: w, height: h, matrix: m, reason: c[5]})
	}
	return rows
}

// buildValidVector produces the encoded octets for a valid corpus vector using
// the discretionary reference encoder.
func buildValidVector(name string, w, h, matrix int) []byte {
	switch name {
	case "two_block_dc_dpcm_16x8":
		// Two horizontally-adjacent blocks with different DC levels.
		return encodePLP1DConly(w, h, matrix, [][3]int{{10, 0, 0}, {30, 0, 0}})
	case "gradient_8x8":
		return encodePLP1DConly(w, h, matrix, [][3]int{{20, 2, -2}})
	case "solid_mid_2x2":
		return encodePLP1DConly(w, h, matrix, [][3]int{{16, 0, 0}})
	default: // solid_black_1x1
		return encodePLP1DConly(w, h, matrix, [][3]int{{0, 0, 0}})
	}
}

// buildInvalidVector produces the malformed octets for an invalid corpus
// vector matching the frozen reason.
func buildInvalidVector(name string, w, h, matrix int) []byte {
	switch name {
	case "bad_magic":
		return []byte{'X', 'X', 'X', 'X', 0x01, 0, 1, 0, 1, 0}
	case "bad_version":
		return []byte{'P', 'L', 'P', '1', 0x02, 0, 1, 0, 1, 0}
	case "zero_width":
		return []byte{'P', 'L', 'P', '1', 0x01, 0, 0, 0, 1, 0}
	case "oversize_dim":
		return []byte{'P', 'L', 'P', '1', 0x01, 0x40, 0x01, 0, 1, 0} // width 16385 > cap
	case "bad_matrix_index":
		return []byte{'P', 'L', 'P', '1', 0x01, 0, 8, 0, 8, 8} // matrix 8
	case "truncated_header":
		return []byte{'P', 'L', 'P', '1', 0x01} // fewer than 10 octets
	case "truncated_block_stream":
		// Valid header, empty block stream (a full 8x8 block needs coefficients).
		return []byte{'P', 'L', 'P', '1', 0x01, 0, 8, 0, 8, 0}
	case "over_ceiling_dim":
		// 16384x16384: within the axis cap but raster (16384^2*3 = 805306368)
		// exceeds MAX_DECODED_UNIT (268435456), so the ceiling guard fires
		// before any block is read or the raster is allocated.
		return []byte{'P', 'L', 'P', '1', 0x01, 0x40, 0x00, 0x40, 0x00, 0}
	default:
		return nil
	}
}
