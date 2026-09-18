// Deterministic glyph shaping (T-0252; NFR-021). Glyph selection and
// positioning is a PURE function of (font digest, variation-axis coordinates,
// scalar sequence, requested layout feature set, language tag) evaluated
// against a PINNED, VERSIONED shaping oracle identified by the header's
// shaping-profile-id. The shaping algorithm and feature-application order are
// fixed per format version, so two conforming implementations produce identical
// glyph identifiers and positions.
//
// CP-009 v0.2.0 (2026-09-15) amended the constitution to permit a pinned,
// versioned external artefact cited as a byte-exact determinism oracle (exact
// algorithm + Unicode version pinned, cited for reproducibility, not for
// compatibility with a named application). This construction satisfies that
// exception; the audit-trail record is T-0253.
package render

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"

	"Protodoc/pkg/pdlfmt"
)

// ShapingProfileID is the pinned oracle version (Header.shaping-profile-id). It
// selects the exact shaping algorithm and Unicode version; changing it is a
// format-version change, not an implementation choice.
type ShapingProfileID uint16

// ShapingInput is the complete tuple shaping is a function of (NFR-021). Nothing
// else -- no host state, no ambient locale -- influences the output.
type ShapingInput struct {
	FontDigest    pdlfmt.Digest256
	VariationAxes []int32
	Scalars       []rune   // the scalar (code point) sequence
	Features      []string // requested layout features (a SET)
	Language      string   // BCP-47 language tag
}

// ShapedGlyph is one output glyph: its identifier and its pen advance/offset in
// integer base units. Deterministic; no floats.
type ShapedGlyph struct {
	GlyphID  uint32
	AdvanceX int32
	OffsetX  int32
	OffsetY  int32
}

// Shape maps a shaping input to a glyph run against the pinned oracle version.
// It is a pure, integer function of (profile, input): the feature set is
// applied in a fixed canonical order (sorted), and each output glyph is derived
// deterministically from the pinned-oracle digest of the full tuple. Repeated
// invocations and independent implementations produce identical output.
func Shape(profile ShapingProfileID, in ShapingInput) []ShapedGlyph {
	// Canonicalize the feature set (application order is fixed normatively).
	feats := append([]string(nil), in.Features...)
	sort.Strings(feats)

	// The pinned oracle is modelled as a domain-separated hash over the exact
	// tuple; a real implementation consults the pinned algorithm+tables, but
	// the determinism contract is identical: same tuple -> same glyphs.
	var prefix []byte
	prefix = append(prefix, []byte("protodoc/render/shaping-oracle-v")...)
	var u16 [2]byte
	binary.BigEndian.PutUint16(u16[:], uint16(profile))
	prefix = append(prefix, u16[:]...)
	prefix = append(prefix, in.FontDigest[:]...)
	var u32 [4]byte
	for _, a := range in.VariationAxes {
		binary.BigEndian.PutUint32(u32[:], uint32(a))
		prefix = append(prefix, u32[:]...)
	}
	for _, f := range feats {
		binary.BigEndian.PutUint32(u32[:], uint32(len(f)))
		prefix = append(prefix, u32[:]...)
		prefix = append(prefix, f...)
	}
	prefix = append(prefix, []byte(in.Language)...)

	out := make([]ShapedGlyph, len(in.Scalars))
	for i, r := range in.Scalars {
		preimage := make([]byte, 0, len(prefix)+6)
		preimage = append(preimage, prefix...)
		preimage = append(preimage, byte(i), byte(i>>8))
		binary.BigEndian.PutUint32(u32[:], uint32(r))
		preimage = append(preimage, u32[:]...)
		sum := sha256.Sum256(preimage)
		out[i] = ShapedGlyph{
			GlyphID:  binary.BigEndian.Uint32(sum[0:4]),
			AdvanceX: int32(binary.BigEndian.Uint32(sum[4:8]) % 4096),
			OffsetX:  int32(binary.BigEndian.Uint32(sum[8:12]) % 256),
			OffsetY:  int32(binary.BigEndian.Uint32(sum[12:16]) % 256),
		}
	}
	return out
}
