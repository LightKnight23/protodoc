# PLP-1 (Protodoc Lossy Profile 1) — frozen decode conformance corpus

Status: FROZEN 2026-09-15 (T-0254). Authored BEFORE any decoder code, per plan.md §10 task (c) and
CP-003's decode-first discipline. Only DECODE is normative (CQ-010); the encoder is writer-discretionary.

This file is the normative reference for the PLP-1 decoder (T-0255). The decoder is bit-exact: for
every valid vector it MUST produce the stated RGB raster octet-for-octet; for every malformed vector it
MUST reject with the stated reason and allocate no proportional memory first.

## Wire format (v1)

All integers big-endian. No floats anywhere.

```
plp1-image      = magic version width height matrix-index block-stream
magic           = %x50 %x4C %x50 %x31          ; "PLP1"
version         = %x01                          ; format version 1
width           = uint16                         ; pixels, 1..PLP1_MAX_DIM
height          = uint16                         ; pixels, 1..PLP1_MAX_DIM
matrix-index    = %x00-07                        ; one of 8 fixed quantization matrices (CQ-010)
block-stream    = *block                          ; ceil(w/8)*ceil(h/8) blocks, raster order, 4:4:4
block           = y-plane cb-plane cr-plane        ; three 8x8 coefficient planes
plane           = dc-dpcm 63(ac-coeff)             ; DC as DPCM from previous block's same-plane DC
dc-dpcm         = svarint                           ; signed minimal varint delta
ac-coeff        = svarint                           ; signed minimal varint, zig-zag order
```

Ceilings:
- PLP1_MAX_DIM = 16384 (per-axis pixel cap; set so a max-square raster exceeds MAX_DECODED_UNIT).
- Decoded raster octets = width*height*3; MUST be bounded against MAX_DECODED_UNIT BEFORE allocation.
- PLP-1 block size fixed 8x8; padding is edge-replicate to a multiple of 8 (data-model.md 690).

Decode pipeline (all integer / fixed-point, bit-exact):
1. Parse header; reject bad magic/version/dim/matrix-index; reject truncation.
2. Bound width*height*3 against MAX_DECODED_UNIT before allocating the raster.
3. For each block in raster order: reconstruct DC by DPCM, read 63 AC in zig-zag, dequantize with the
   selected fixed matrix, inverse-zig-zag, apply the fixed integer 8x8 IDCT (integer matrix + fixed
   shift schedule), producing an 8x8 sample block per plane.
4. Convert YCbCr->RGB with BT.601 scaled-integer coefficients and round-half-up, clamp 0..255.
5. Crop the edge-replicated padding back to width x height.

## Conformance vectors

Each vector is `<name>.plp1` (input octets) plus, for VALID vectors, `<name>.rgb` (expected raster:
width*height*3 octets) and, for INVALID vectors, an entry in `manifest.tsv` naming the rejection reason.

`manifest.tsv` columns: name<TAB>kind(valid|invalid)<TAB>width<TAB>height<TAB>matrix<TAB>reason.

Frozen vector set (syntactic decode coverage):

| name | kind | why |
|---|---|---|
| solid_black_1x1 | valid | minimal 1x1, all-zero coefficients -> black |
| solid_mid_2x2 | valid | sub-block image, edge-replicate padding to 8x8 |
| gradient_8x8 | valid | one full block with non-zero AC coefficients |
| two_block_dc_dpcm_16x8 | valid | DC-DPCM across two horizontally-adjacent blocks |
| bad_magic | invalid | first four octets are not "PLP1" |
| bad_version | invalid | version octet != 1 |
| zero_width | invalid | width = 0 |
| oversize_dim | invalid | width > PLP1_MAX_DIM |
| bad_matrix_index | invalid | matrix-index = 8 (only 0..7 valid) |
| truncated_header | invalid | fewer than the fixed header octets |
| truncated_block_stream | invalid | header ok but block stream ends mid-coefficient |
| over_ceiling_dim | invalid | width*height*3 exceeds MAX_DECODED_UNIT -> reject before alloc |

The `.plp1`/`.rgb` octets are generated deterministically by the frozen reference encoder embedded in the
T-0255 test (writer-discretionary encoder, decode-bit-exact); the corpus is the set of names+kinds+reasons
frozen here. T-0255 asserts the decoder reproduces every valid `.rgb` bit-exact and rejects every invalid
vector with the named reason, allocating no proportional memory for the over-ceiling case.
